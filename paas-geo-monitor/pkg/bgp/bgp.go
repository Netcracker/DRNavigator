package bgp

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/netcracker/drnavigator/paas-geo-monitor/logger"
	"github.com/prometheus/client_golang/prometheus"

	v3 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	"github.com/projectcalico/api/pkg/client/clientset_generated/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type BGPMetrics struct {
	BgpPeer  *prometheus.GaugeVec
	BgpRoute *prometheus.GaugeVec
}

var BgpMetrics = &BGPMetrics{
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

func GetClientSet() (clientSet *clientset.Clientset) {

	var (
		kubeconfig *rest.Config
	)

	// Use KUBECONFIG env variable to try debug code locally
	if os.Getenv("KUBECONFIG") != "" {

		config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
		if err != nil {
			// panic(err.Error())
			errorString := fmt.Sprintf("Can't parse KUBECONFIG file: %s", err)
			panic(errorString)
		}
		kubeconfig = config

	} else {

		config, err := rest.InClusterConfig()
		if err != nil {
			errorString := fmt.Sprintf("Can't use InClusterConfig: %s", err)
			panic(errorString)
		}
		kubeconfig = config

	}

	clientSet, err := clientset.NewForConfig(kubeconfig)

	if err != nil {
		panic(err.Error())
	}

	return clientSet
}

func GetCrStatus(clientSet *clientset.Clientset) (list *v3.CalicoNodeStatusList, err error) {

	// List Calico Node Statuses.
	list, err = clientSet.ProjectcalicoV3().CalicoNodeStatuses().List(context.Background(), v1.ListOptions{})

	if err != nil {
		// panic(err.Error())
		return list, err
	}

	return list, nil
}

func UpdateBgpMetrics(bgpMetrics *BGPMetrics, list *v3.CalicoNodeStatusList, paasBgpCheckTimeout int) {

	var (
		peer_status  float64
		route_status float64
	)

	log := logger.SimpleLogger()

	bgpMetrics.BgpPeer.Reset()
	bgpMetrics.BgpRoute.Reset()

	for _, item := range list.Items {
		for _, peer := range item.Status.BGP.PeersV4 {
			if peer.Type == "GlobalPeer" {
				if peer.State == "Established" {
					peer_status = 1
				} else {
					peer_status = 0
				}

				if item.Status.LastUpdated.Unix() < time.Now().Unix()-int64(paasBgpCheckTimeout) {
					peer_status = -1
				}

				bgpMetrics.BgpPeer.With(prometheus.Labels{
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

				if item.Status.LastUpdated.Unix() < time.Now().Unix()-int64(paasBgpCheckTimeout) {
					route_status = -1
				}

				bgpMetrics.BgpRoute.With(prometheus.Labels{
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
