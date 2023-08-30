package bgp

import (
	"context"
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
	bgpPeer  *prometheus.GaugeVec
	bgpRoute *prometheus.GaugeVec
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

func getCrStatus(clientSet *clientset.Clientset) (list *v3.CalicoNodeStatusList) {

	// List Calico Node Statuses.
	list, err := clientSet.ProjectcalicoV3().CalicoNodeStatuses().List(context.Background(), v1.ListOptions{})

	if err != nil {
		panic(err.Error())
	}

	return list
}

func updateBgpMetrics(bgpMetrics *BGPMetrics, list *v3.CalicoNodeStatusList) {

	var (
		peer_status  float64
		route_status float64
	)

	log := logger.SimpleLogger()

	// bgpMetrics.bgpPeer.Reset()
	// bgpMetrics.bgpRoute.Reset()

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

	// fmt.Printf("bgpMetrics.bgpPeer: %v\n", bgpMetrics.bgpPeer)
	// fmt.Printf("bgpMetrics.bgpRoute: %v\n", bgpMetrics.bgpRoute)
}
