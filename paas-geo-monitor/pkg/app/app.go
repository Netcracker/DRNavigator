package app

import (
	"context"
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
	"github.com/netcracker/drnavigator/paas-geo-monitor/pkg/client"
	"github.com/netcracker/drnavigator/paas-geo-monitor/pkg/resources"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v3"

	v3 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	"github.com/projectcalico/api/pkg/client/clientset_generated/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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

type BGPMetrics struct {
	bgpPeer  *prometheus.GaugeVec
	bgpRoute *prometheus.GaugeVec
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

	bgpMetrics := &BGPMetrics{
		prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "paas_geo_monitor_bgp_peer",
				Help: "paas_geo_monitor bgp global peer: : 1 - Established, 0 - not Established, -1 - did not update",
			},
			[]string{"node", "peer_ip", "state", "type"},
		),
		prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "paas_geo_monitor_bgp_route",
				Help: "paas_geo_monitor bgp route: 1 - existed, 0 - not existed, -1 - did not update",
			},
			[]string{"node", "destination", "source_type", "peer_ip", "type", "interface"},
		),
	}

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
	if err := prometheus.Register(bgpMetrics.bgpPeer); err != nil {
		return fmt.Errorf("Can't regist bgp_peer prometheus metric: %s", err)
	}
	if err := prometheus.Register(bgpMetrics.bgpRoute); err != nil {
		return fmt.Errorf("Can't regist bgp_route prometheus metric: %s", err)
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
		go func() {
			clientSet := getClientSet()
			for {
				calicoStatusList := getCRStatus(clientSet)
				updateBGPMetrics(bgpMetrics, calicoStatusList)
				time.Sleep(10 * time.Second)
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

func getClientSet() (clientSet *clientset.Clientset) {

	var (
		kubeconfig *rest.Config
	)

	// Use KUBECONFIG env variable to try debug code locally
	if os.Getenv("KUBECONFIG") != "" {

		config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
		if err != nil {
			panic(err.Error())
		}
		kubeconfig = config

	} else {

		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		kubeconfig = config

	}

	clientSet, err := clientset.NewForConfig(kubeconfig)

	if err != nil {
		panic(err.Error())
	}

	return clientSet
}

func getCRStatus(clientSet *clientset.Clientset) (list *v3.CalicoNodeStatusList) {

	// List Calico Node Statuses.
	list, err := clientSet.ProjectcalicoV3().CalicoNodeStatuses().List(context.Background(), v1.ListOptions{})

	if err != nil {
		panic(err.Error())
	}

	return list
}

func updateBGPMetrics(bgpMetrics *BGPMetrics, list *v3.CalicoNodeStatusList) {

	var (
		peer_status  float64
		route_status float64
	)

	log := logger.SimpleLogger()

	for _, item := range list.Items {
		for _, peer := range item.Status.BGP.PeersV4 {
			if peer.Type == "GlobalPeer" {
				if peer.State == "Established" {
					peer_status = 1
				} else {
					peer_status = 0
				}

				if item.Status.LastUpdated.Unix() < time.Now().Unix()-30 {
					peer_status = -1
				}

				bgpMetrics.bgpPeer.With(prometheus.Labels{
					"node":    item.Spec.Node,
					"peer_ip": peer.PeerIP,
					"state":   string(peer.State),
					"type":    string(peer.Type),
				}).Set(peer_status)

				log.Debugf("paas_geo_monitor_bgp_peer{node=%#v, peer_ip=%#v, state=%#v, type=%#v}",
					item.Spec.Node,
					peer.PeerIP,
					peer.State,
					peer.Type)
			}
		}

		for _, route := range item.Status.Routes.RoutesV4 {
			if route.Type == "FIB" && route.LearnedFrom.SourceType == "BGPPeer" {

				route_status = 1

				if item.Status.LastUpdated.Unix() < time.Now().Unix()-30 {
					route_status = -1
				}

				bgpMetrics.bgpRoute.With(prometheus.Labels{
					"node":        item.Spec.Node,
					"destination": route.Destination,
					"source_type": string(route.LearnedFrom.SourceType),
					"peer_ip":     route.LearnedFrom.PeerIP,
					"type":        string(route.Type),
					"interface":   route.Interface,
				}).Set(route_status)

				log.Debugf("paas_geo_monitor_bgp_route{node=%#v, destination=%#v, source_type=%#v, peer_ip=%#v, type=%#v, inetface=%#v}",
					item.Spec.Node,
					route.Destination,
					route.LearnedFrom.SourceType,
					route.LearnedFrom.PeerIP,
					route.Type,
					route.Interface)
			}
		}
	}
}
