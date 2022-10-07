package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HttpClient struct {
}

func (h *HttpClient) Resolve(name string) (string, error) {
	ips, err := net.LookupIP(name)
	if err != nil {
		return "", fmt.Errorf("unable to resolve name %s: %w", name, err)
	}

	// currently we expect only ClusterIP services which have only one IP address
	if len(ips) != 1 {
		return "", fmt.Errorf("expected one resolved IP for name %s, but got: %d IP addresses", name, len(ips))
	}

	return ips[0].String(), nil
}

func (h *HttpClient) Get(reqUrl string, ip string) (string, error) {
	// TODO support insecure/CA
	client := &http.Client{Transport: &http.Transport{}, Timeout: 5 * time.Second}

	if ip != "" {
		client.Transport.(*http.Transport).DialContext = func(c context.Context, n, addr string) (net.Conn, error) {
			probeUrl, err := url.ParseRequestURI(reqUrl)
			if err != nil {
				return nil, fmt.Errorf("failed to parse probe url: %v", err)
			}

			// we expect that addr will have hostname before ":", and port after ":"
			addrParts := strings.Split(addr, ":")
			if addrParts[0] == probeUrl.Hostname() {
				addr = fmt.Sprintf("%s:%s", ip, addrParts[1])
			}
			return (&net.Dialer{}).DialContext(c, n, addr)
		}
	}

	res, err := client.Get(reqUrl)
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
