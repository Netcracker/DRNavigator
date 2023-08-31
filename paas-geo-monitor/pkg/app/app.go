package app

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/netcracker/drnavigator/paas-geo-monitor/logger"
	"github.com/netcracker/drnavigator/paas-geo-monitor/pkg/bgp"
	"github.com/netcracker/drnavigator/paas-geo-monitor/pkg/client"
	"github.com/netcracker/drnavigator/paas-geo-monitor/pkg/resources"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Port  int
	Peers []resources.Peer
}

type PeersMetrics struct {
	peersDnsStatus *prometheus.GaugeVec
	peersSvcStatus *prometheus.GaugeVec
	peersPodStatus *prometheus.GaugeVec
}

func Serve(cfg *Config) error {
	e := echo.New()
	e.Use(middleware.Logger())

	// Create custom metrics
	paasHealth := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "paas_geo_monitor_health",
			Help: "paas-geo-monitor pod health",
		},
	)

	peersMetrics := &PeersMetrics{
		prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "peer_dns_status",
				Help: "Peer dns status: 1 - resolved, 0 - not resolved, -1 - can't check connection",
			},
			[]string{"peer_name"},
		),
		prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "peer_svc_status",
				Help: "Peer svc status: 1 - available, 0 - not available, -1 - can't check connection",
			},
			[]string{"peer_name"},
		),
		prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "peer_pod_status",
				Help: "Peer pod status: 1 - available, 0 - not available, -1 - can't check connection",
			},
			[]string{"peer_name"},
		),
	}

	// Regist custom metrics
	if err := prometheus.Register(paasHealth); err != nil {
		return fmt.Errorf("Can't regist paas_health prometheus metric: %s", err)
	}
	if err := prometheus.Register(peersMetrics.peersDnsStatus); err != nil {
		return fmt.Errorf("Can't regist peer_dns_status prometheus metric: %s", err)
	}
	if err := prometheus.Register(peersMetrics.peersSvcStatus); err != nil {
		return fmt.Errorf("Can't regist peers_svc_status prometheus metric: %s", err)
	}
	if err := prometheus.Register(peersMetrics.peersPodStatus); err != nil {
		return fmt.Errorf("Can't regist peers_pod_status prometheus metric: %s", err)
	}

	pingIp := os.Getenv("PING_IP")
	if net.ParseIP(pingIp) == nil {
		return fmt.Errorf("incorrect or empty PING_IP: '%s'", pingIp)
	}
	paasPingPeers := true
	var err error
	if paasPingPeersEnv, exist := os.LookupEnv("PAAS_PING_PEERS"); exist {
		paasPingPeers, err = strconv.ParseBool(paasPingPeersEnv)
		if err != nil {
			return fmt.Errorf("Can't parse PAAS_PING_PEERS value: %s", err)
		}
	}
	paasPingTime := 5
	if paasPingTimeEnv, exist := os.LookupEnv("PAAS_PING_TIME"); exist {
		paasPingTime, err = strconv.Atoi(paasPingTimeEnv)
		if err != nil {
			return fmt.Errorf("Can't parse PAAS_PING_TIME value: %s", err)
		}
	}

	e.Use(echoprometheus.NewMiddleware("paas_geo_monitor"))
	e.GET("/metrics", echoprometheus.NewHandler())
	e.GET("/ping", pingHandler(pingIp, paasHealth))
	e.GET("/peers/status", getPeersStatusHandler(cfg.Peers))

	// Ping peer status in separate threads
	if paasPingPeers {
		for i := range cfg.Peers {
			go pingPeersStatus(cfg.Peers[i], peersMetrics, paasPingTime)
		}
	}

	if os.Getenv("PAAS_BGP_METRICS") == "true" {

		if err := prometheus.Register(bgp.BgpMetrics.BgpPeer); err != nil {
			return fmt.Errorf("Can't regist bgp_peer prometheus metric: %s", err)
		}
		if err := prometheus.Register(bgp.BgpMetrics.BgpRoute); err != nil {
			return fmt.Errorf("Can't regist bgp_route prometheus metric: %s", err)
		}

		paasBgpCheckPeriod := 10
		if paasBgpCheckPeriodEnv, exist := os.LookupEnv("PAAS_BGP_CHECK_PERIOD"); exist {
			paasBgpCheckPeriod, err = strconv.Atoi(paasBgpCheckPeriodEnv)
			if err != nil {
				return fmt.Errorf("Can't parse PAAS_BGP_CHECK_PERIOD: %s", err)
			}
		}

		paasBgpCheckTimeout := 30
		if paasBgpCheckTimeoutEnv, exist := os.LookupEnv("PAAS_BGP_CHECK_TIMEOUT"); exist {
			paasBgpCheckTimeout, err = strconv.Atoi(paasBgpCheckTimeoutEnv)
			if err != nil {
				return fmt.Errorf("Can't parse PAAS_BGP_CHECK_TIMEOUT value: %s \n", err)
			}
		}

		go func() {
			clientSet := bgp.GetClientSet()
			for {
				calicoStatusList := bgp.GetCrStatus(clientSet)
				bgp.UpdateBgpMetrics(bgp.BgpMetrics, calicoStatusList, paasBgpCheckTimeout)
				time.Sleep(time.Duration(paasBgpCheckPeriod) * time.Second)
			}
		}()
	}

	// todo: support TLS
	return e.Start(fmt.Sprintf(":%d", cfg.Port))
}

func GetConfig(cfgPath string) (*Config, error) {
	file, err := os.Open(cfgPath)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(cfg)
	if err != nil {
		return nil, err
	}

	for i := range cfg.Peers {
		for j := i + 1; j < len(cfg.Peers); j++ {
			if cfg.Peers[i].Name == cfg.Peers[j].Name {
				return nil, errors.New("Found more than one peers with name " + cfg.Peers[i].Name + ". Peer name should be unique")
			}
		}
		err := cfg.Peers[i].Init(&client.HttpClient{})
		if err != nil {
			return nil, err
		}
	}

	if cfg.Port == 0 {
		cfg.Port = 8080
	}

	return cfg, nil
}

func pingHandler(pingIp string, paasHealth prometheus.Gauge) func(c echo.Context) error {
	return func(c echo.Context) error {
		paasHealth.Set(1)
		return c.String(http.StatusOK, pingIp)
	}
}

func getPeersStatusHandler(peers []resources.Peer) func(c echo.Context) error {
	return func(c echo.Context) error {
		statuses := make([]*resources.PeerStatus, len(peers))
		for i := range peers {
			s, err := peers[i].Status()
			if err != nil {
				return fmt.Errorf("failed to collect statuses: %s", err)
			}
			statuses[i] = s
		}

		resp, err := yaml.Marshal(statuses)
		if err != nil {
			return fmt.Errorf("failed to marshal response: %s", err)
		}
		return c.String(http.StatusOK, fmt.Sprintf("%v", string(resp)))
	}
}

func pingPeersStatus(peer resources.Peer, peersMetrics *PeersMetrics, pingTime int) {
	log := logger.SimpleLogger()
	for {
		log.Debugf("[Peer %s] Ping status peer started", peer.Name)
		dnsStatus := -1
		svcStatus := -1
		podStatus := -1
		s, err := peer.Status()
		if err == nil {
			if s.ClusterIpStatus.DnsStatus.Resolved {
				dnsStatus = 1
				if s.ClusterIpStatus.SvcStatus.Available {
					svcStatus = 1
					if s.ClusterIpStatus.PodStatus.Available {
						podStatus = 1
					} else {
						podStatus = 0
						log.Warnf("[Peer %s]  Pod2pod connection fails: %s", peer.Name, s.ClusterIpStatus.PodStatus.Error)
					}
				} else {
					svcStatus = 0
					log.Warnf("[Peer %s] Pod2service connection fails: %s", peer.Name, s.ClusterIpStatus.SvcStatus.Error)
				}
			} else {
				dnsStatus = 0
				log.Warnf("[Peer %s] Can't resolve dns status: %s", peer.Name, s.ClusterIpStatus.DnsStatus.Error)
			}
		} else {
			log.Errorf("[Peer %s] Can't check status: %s", peer.Name, err)
		}
		peersMetrics.peersDnsStatus.WithLabelValues(peer.Name).Set(float64(dnsStatus))
		peersMetrics.peersSvcStatus.WithLabelValues(peer.Name).Set(float64(svcStatus))
		peersMetrics.peersPodStatus.WithLabelValues(peer.Name).Set(float64(podStatus))
		log.Debugf("[Peer %s] Ping status peer finished, sleep %ds", peer.Name, pingTime)
		time.Sleep(time.Duration(pingTime) * time.Second)
	}
}
