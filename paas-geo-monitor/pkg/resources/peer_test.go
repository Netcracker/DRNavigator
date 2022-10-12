package resources_test

import (
	"errors"
	"github.com/netcracker/drnavigator/paas-geo-monitor/pkg/resources"
	"testing"
)

func TestPeer_StatusCorrect(t *testing.T) {
	p := resources.Peer{Name: "peer-test", ClusterIp: resources.ClusterIp{Name: "test.com"}}
	cMock := &ClientMock{
		ResolvedIp: "1.1.1.1",
		PingIpMapping: map[string]string{
			"1.1.1.1": "2.2.2.2",
			"2.2.2.2": "2.2.2.2",
		},
	}
	err := p.Init(cMock)
	if err != nil {
		t.Fatalf("failed to init peer: %s", err)
	}

	status, err := p.Status()
	if err != nil {
		t.Fatalf("failed to get status: %s", err)
	}

	if status.ClusterIpStatus.DnsStatus.Resolved == false {
		t.Fatalf("expected to resolve DNS, but got: %s", status.ClusterIpStatus.DnsStatus.Error)
	}
	if status.ClusterIpStatus.SvcStatus.Available == false {
		t.Fatalf("expected to reach svc, but got: %s", status.ClusterIpStatus.SvcStatus.Error)
	}
	if status.ClusterIpStatus.PodStatus.Available == false {
		t.Fatalf("expected to reach pod, but got: %s", status.ClusterIpStatus.PodStatus.Error)
	}
}

func TestPeer_DnsFailed(t *testing.T) {
	p := resources.Peer{Name: "peer-test", ClusterIp: resources.ClusterIp{Name: "test.com"}}
	cMock := &ClientMock{
		ResolvedIp: "",
	}
	err := p.Init(cMock)
	if err != nil {
		t.Fatalf("failed to init peer: %s", err)
	}

	status, err := p.Status()
	if err != nil {
		t.Fatalf("failed to get status: %s", err)
	}

	if status.ClusterIpStatus.DnsStatus.Resolved == true {
		t.Fatalf("expected to fail DNS, but got: %s", status.ClusterIpStatus.SvcStatus.Address)
	}
}

func TestPeer_PingFailed(t *testing.T) {
	p := resources.Peer{Name: "peer-test", ClusterIp: resources.ClusterIp{Name: "test.com"}}
	cMock := &ClientMock{
		ResolvedIp: "1.1.1.1",
	}
	err := p.Init(cMock)
	if err != nil {
		t.Fatalf("failed to init peer: %s", err)
	}

	status, err := p.Status()
	if err != nil {
		t.Fatalf("failed to get status: %s", err)
	}

	if status.ClusterIpStatus.DnsStatus.Resolved == false {
		t.Fatalf("expected to resolve DNS, but got: %s", status.ClusterIpStatus.DnsStatus.Error)
	}
	if status.ClusterIpStatus.SvcStatus.Available == true {
		t.Fatalf("expected svc not reachable, but got response: %s", status.ClusterIpStatus.PodStatus.Address)
	}
}

type ClientMock struct {
	ResolvedIp    string
	PingIpMapping map[string]string
}

func (c *ClientMock) Resolve(name string) (string, error) {
	if c.ResolvedIp != "" {
		return c.ResolvedIp, nil
	}

	return "", errors.New("no such host")
}

func (c *ClientMock) Get(url string, ip string) (string, error) {
	if resp, ok := c.PingIpMapping[ip]; ok {
		return resp, nil
	}

	return "", errors.New("not reachable")
}
