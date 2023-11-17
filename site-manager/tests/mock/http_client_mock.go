package mock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/netcracker/drnavigator/site-manager/pkg/model"
)

type HttpClientMock struct {
	ServiceStatus model.ServiceSiteManagerResponse
	ServiceHealth model.ServiceHealthzResponse
}

func (hcm *HttpClientMock) Do(req *http.Request) (*http.Response, error) {
	switch req.URL.Path {
	case "/sitemanager":
		byteData, _ := json.Marshal(hcm.ServiceStatus)
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader(byteData)),
		}, nil
	case "/health":
		byteData, _ := json.Marshal(hcm.ServiceHealth)
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader(byteData)),
		}, nil
	default:
		return nil, fmt.Errorf("http mock does not support path %s", req.URL.Path)
	}
}
