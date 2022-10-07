package app

import (
	"fmt"
	"git.netcracker.com/prod.platform.ha/paas-geo-monitor/pkg/client"
	"git.netcracker.com/prod.platform.ha/paas-geo-monitor/pkg/resources"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gopkg.in/yaml.v3"
	"net"
	"net/http"
	"os"
)

type Config struct {
	Port      int
	Neighbors []resources.Neighbor
}

func Serve(cfg *Config) error {
	e := echo.New()
	e.Use(middleware.Logger())

	pingIp := os.Getenv("PING_IP")
	if net.ParseIP(pingIp) == nil {
		return fmt.Errorf("incorrect or empty PING_IP: '%s'", pingIp)
	}

	e.GET("/ping", pingHandler(pingIp))
	e.GET("/neighbors/status", getNeighborsStatusHandler(cfg.Neighbors))

	port := cfg.Port
	if port == 0 {
		port = 8080
	}

	// todo: support TLS
	return e.Start(fmt.Sprintf(":%d", port))
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

	for i := range cfg.Neighbors {
		err := cfg.Neighbors[i].Init(&client.HttpClient{})
		if err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func pingHandler(pingIp string) func(c echo.Context) error {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, pingIp)
	}
}

func getNeighborsStatusHandler(neighbors []resources.Neighbor) func(c echo.Context) error {
	return func(c echo.Context) error {
		statuses := make([]*resources.NeighborStatus, len(neighbors))
		for i := range neighbors {
			s, err := neighbors[i].Status()
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
