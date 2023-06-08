package app

import (
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
	peers_dns_status *prometheus.GaugeVec
	peers_svc_status *prometheus.GaugeVec
	peers_pod_status *prometheus.GaugeVec
}

func Serve(cfg *Config) error {
	e := echo.New()
	e.Use(middleware.Logger())

	// Create custom metrics
	paas_health := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "paas_geo_monitor_health",
			Help: "paas-geo-monitor pod health",
		},
	)
	peers_metrics := &PeersMetrics{
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
	if err := prometheus.Register(paas_health); err != nil {
		return fmt.Errorf("Can't regist paas_health prometheus metric: %s", err)
	}
	if err := prometheus.Register(peers_metrics.peers_dns_status); err != nil {
		return fmt.Errorf("Can't regist peer_dns_status prometheus metric: %s", err)
	}
	if err := prometheus.Register(peers_metrics.peers_svc_status); err != nil {
		return fmt.Errorf("Can't regist peers_svc_status prometheus metric: %s", err)
	}
	if err := prometheus.Register(peers_metrics.peers_pod_status); err != nil {
		return fmt.Errorf("Can't regist peers_pod_status prometheus metric: %s", err)
	}

	pingIp := os.Getenv("PING_IP")
	if net.ParseIP(pingIp) == nil {
		return fmt.Errorf("incorrect or empty PING_IP: '%s'", pingIp)
	}
	paas_ping_peers := true
	var err error
	if paas_ping_peers_env, exist := os.LookupEnv("PAAS_PING_PEERS"); exist {
		paas_ping_peers, err = strconv.ParseBool(paas_ping_peers_env)
		if err != nil {
			return fmt.Errorf("Can't parse PAAS_PING_PEERS value: %s", err)
		}
	}
	paas_ping_time := 5
	if paas_ping_time_env, exist := os.LookupEnv("PAAS_PING_TIME"); exist {
		paas_ping_time, err = strconv.Atoi(paas_ping_time_env)
		if err != nil {
			return fmt.Errorf("Can't parse PAAS_PING_TIME value: %s", err)
		}
	}

	e.Use(echoprometheus.NewMiddleware("paas_geo_monitor"))
	e.GET("/metrics", echoprometheus.NewHandler())
	e.GET("/ping", pingHandler(pingIp, paas_health))
	e.GET("/peers/status", getPeersStatusHandler(cfg.Peers))

	// Ping peer status in separate threads
	if paas_ping_peers {
		for i := range cfg.Peers {
			go pingPeersStatus(cfg.Peers[i], peers_metrics, paas_ping_time)
		}
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

func pingHandler(pingIp string, paas_health prometheus.Gauge) func(c echo.Context) error {
	return func(c echo.Context) error {
		paas_health.Set(1)
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

func pingPeersStatus(peer resources.Peer, peers_metrics *PeersMetrics, ping_time int) {
	log := logger.SimpleLogger()
	for {
		log.Debugf("[Peer %s] Ping status peer started", peer.Name)
		dns_status := -1
		svc_status := -1
		pod_status := -1
		s, err := peer.Status()
		if err == nil {
			if s.ClusterIpStatus.DnsStatus.Resolved {
				dns_status = 1
				if s.ClusterIpStatus.SvcStatus.Available {
					svc_status = 1
					if s.ClusterIpStatus.PodStatus.Available {
						pod_status = 1
					} else {
						pod_status = 0
						log.Debugf("[Peer %s]  Pod2pod connection fails: %s", peer.Name, s.ClusterIpStatus.PodStatus.Error)
					}
				} else {
					svc_status = 0
					log.Debugf("[Peer %s] Pod2service connection fails: %s", peer.Name, s.ClusterIpStatus.SvcStatus.Error)
				}
			} else {
				dns_status = 0
				log.Debugf("[Peer %s] Can't resolve dns status: %s", peer.Name, s.ClusterIpStatus.DnsStatus.Error)
			}
		} else {
			log.Errorf("[Peer %s] Can't check status: %s", peer.Name, err)
		}
		peers_metrics.peers_dns_status.WithLabelValues(peer.Name).Set(float64(dns_status))
		peers_metrics.peers_svc_status.WithLabelValues(peer.Name).Set(float64(svc_status))
		peers_metrics.peers_pod_status.WithLabelValues(peer.Name).Set(float64(pod_status))
		log.Debugf("[Peer %s] Ping status peer finished, sleep %ds", peer.Name, ping_time)
		time.Sleep(time.Duration(ping_time) * time.Second)
	}
}
