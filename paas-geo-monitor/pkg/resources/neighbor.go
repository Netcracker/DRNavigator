package resources

import (
	"fmt"
	"net"
	"net/url"
)

// Neighbor represents monitor instance from another cluster
type Neighbor struct {
	Name      string
	ClusterIp ClusterIp `yaml:"clusterIp"`

	Client NeighborClient
}

// ClusterIp represents a k8s ClusterIP service
type ClusterIp struct {
	Name     string
	SvcPort  int `yaml:"svcPort,omitempty"`
	PodPort  int `yaml:"podPort,omitempty"`
	Protocol string
}

// NeighborStatus represents a status of another monitor instance from the perspective of current instance
type NeighborStatus struct {
	Name            string
	ClusterIpStatus *ClusterIpStatus `yaml:"clusterIpStatus"`
}

// ClusterIpStatus represents a status of k8s ClusterIp service
type ClusterIpStatus struct {
	ClusterIp `yaml:"clusterIp"`
	DnsStatus *DnsStatus `yaml:"dnsStatus,omitempty"`
	SvcStatus *SvcStatus `yaml:"svcStatus,omitempty"`
	PodStatus *PodStatus `yaml:"podStatus,omitempty"`
}

// DnsStatus represents dns status of k8s ClusterIp service (resolved or not)
type DnsStatus struct {
	Resolved bool
	Error    string `yaml:"error,omitempty"`
}

// SvcStatus represents address status of k8s ClusterIp service (available or not)
type SvcStatus struct {
	Available bool
	Address   string `yaml:"address,omitempty"`
	Error     string `yaml:"error,omitempty"`
}

// PodStatus represents address status of k8s pod (available or not)
type PodStatus struct {
	Available bool
	Address   string `yaml:"address,omitempty"`
	Error     string `yaml:"error,omitempty"`
}

type NeighborClient interface {
	Resolve(name string) (string, error)
	Get(url string, ip string) (string, error)
}

func (n *Neighbor) Init(client NeighborClient) error {
	n.Client = client
	if n.ClusterIp.PodPort == 0 {
		n.ClusterIp.PodPort = 8080
	}
	if n.ClusterIp.SvcPort == 0 {
		n.ClusterIp.SvcPort = 8080
	}
	if n.ClusterIp.Protocol == "" {
		n.ClusterIp.Protocol = "http"
	}

	reqUrl := fmt.Sprintf("%s://%s/ping", n.ClusterIp.Protocol, n.ClusterIp.Name)
	_, err := url.ParseRequestURI(reqUrl)
	if err != nil {
		return fmt.Errorf("error validating %s neighbor, incorrect url: %w", n.Name, err)
	}

	return nil
}

func (n *Neighbor) Status() (*NeighborStatus, error) {
	if n.Client == nil {
		return nil, fmt.Errorf("not initialized with a client")
	}

	clusterIpStatus := ClusterIpStatus{ClusterIp: n.ClusterIp, DnsStatus: &DnsStatus{}}
	neighborStatus := &NeighborStatus{Name: n.Name, ClusterIpStatus: &clusterIpStatus}

	resolvedSvcIp, err := n.Client.Resolve(n.ClusterIp.Name)
	if err != nil {
		clusterIpStatus.DnsStatus.Error = fmt.Sprintf("failed to resolve neighbor: %s", err)
		return neighborStatus, nil
	}
	clusterIpStatus.DnsStatus.Resolved = true

	clusterIpStatus.SvcStatus = &SvcStatus{Address: resolvedSvcIp}
	svcPingUrl := fmt.Sprintf("%s://%s:%d/ping", n.ClusterIp.Protocol, n.ClusterIp.Name, n.ClusterIp.SvcPort)
	svcPingResponse, err := n.Client.Get(svcPingUrl, resolvedSvcIp)
	if err != nil {
		clusterIpStatus.SvcStatus.Error = fmt.Sprintf("failed to ping neighbor service: %s", err)
		return neighborStatus, nil
	}
	// svcPingResponse should be a valid IP of the pod
	if net.ParseIP(svcPingResponse) == nil {
		clusterIpStatus.SvcStatus.Error = fmt.Sprintf("neighbor ping returned incorrect IP: %s", err)
		return neighborStatus, nil
	}
	clusterIpStatus.SvcStatus.Available = true

	clusterIpStatus.PodStatus = &PodStatus{Address: svcPingResponse}
	podPingUrl := fmt.Sprintf("%s://%s:%d/ping", n.ClusterIp.Protocol, n.ClusterIp.Name, n.ClusterIp.PodPort)
	podPingResponse, err := n.Client.Get(podPingUrl, svcPingResponse)
	if err != nil {
		clusterIpStatus.PodStatus.Error = fmt.Sprintf("failed to ping neighbor pod: %s", err)
		return neighborStatus, nil
	}
	// both svc and pod pings should return the same IP
	if svcPingResponse != podPingResponse {
		clusterIpStatus.PodStatus.Error = fmt.Sprintf("expected neighbor pod IP %s, got %s", svcPingResponse, podPingResponse)
		return neighborStatus, nil
	}
	clusterIpStatus.PodStatus.Available = true

	return neighborStatus, nil
}
