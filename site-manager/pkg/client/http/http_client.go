package http_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	ctrl "sigs.k8s.io/controller-runtime"
)

var httpLog = ctrl.Log.WithName("http-client")

// HttpClientInterface is an interface to do http requests
type HttpClientInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

// DoGetRequest does GET request to given url using given http client, authorization properties, retry count and stores result in given obj
// return the status code of request and error
func DoGetRequest[V any](client HttpClientInterface, url string, token string, useAuth bool, retry int, obj *V) (int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	if useAuth {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	httpLog.V(1).Info("REST request", "url", url, "type", "GET")

	var response *http.Response

	for retry > 0 {
		response, err = client.Do(req)
		if err != nil {
			httpLog.Error(err, "Request error", "url", url, "type", "GET")
			retry -= 1
		} else {
			break
		}
	}

	if response != nil {
		defer response.Body.Close()
		textData, err := io.ReadAll(response.Body)
		if err != nil {
			httpLog.Error(err, "Can't retreive response body", "url", url, "type", "GET")
			return response.StatusCode, err
		}
		httpLog.V(1).Info("Request is done", "url", url, "type", "GET", "status", response.Status, "response", string(textData))
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			httpLog.Error(err, "Unexpected status code", "url", url, "type", "GET", "status_code", response.StatusCode)
			return response.StatusCode, fmt.Errorf("unexpected status code %d, response: %s", response.StatusCode, textData)
		} 
		decoder := json.NewDecoder(bytes.NewReader(textData))
		//decoder.DisallowUnknownFields()
		if err := decoder.Decode(&obj); err != nil {
			httpLog.Error(err, "Wrong JSON data received", "url", url, "type", "GET")
			return response.StatusCode, fmt.Errorf("wrong json data received: %s", textData)
		}
		return response.StatusCode, nil
	}
	return 0, err
}

// DoPostRequest does POST request to given url using given http client, authorization properties, retry count and stores result in given obj
// return the status code of request and error
func DoPostRequest[V any, T any](client HttpClientInterface, url string, bodyObj *T, token string, useAuth bool, retry int, obj *V) (int, error) {
	byteBody, err := json.Marshal(bodyObj)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(byteBody))
	if err != nil {
		return 0, err
	}
	if useAuth {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	req.Header.Set("Content-Type", "application/json")

	httpLog.V(1).Info("REST request", "url", url, "type", "POST", "body", string(byteBody))

	var response *http.Response

	for retry > 0 {
		response, err = client.Do(req)
		if err != nil {
			httpLog.Error(err, "Request error", "url", url, "type", "POST")
			retry -= 1
		} else {
			break
		}
	}

	if response != nil {
		defer response.Body.Close()
		textData, err := io.ReadAll(response.Body)
		if err != nil {
			httpLog.Error(err, "Can't retreive response body", "url", url, "type", "POST")
			return response.StatusCode, err
		}
		httpLog.V(1).Info("Request is done", "status", "url", url, "type", "POST", response.Status, "response", string(textData))
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			httpLog.Error(err, "Unexpected status code", "url", url, "type", "GET", "status_code", response.StatusCode)
			return response.StatusCode, fmt.Errorf("unexpected status code %d, response: %s", response.StatusCode, textData)
		} 
		decoder := json.NewDecoder(bytes.NewReader(textData))
		//decoder.DisallowUnknownFields()
		if err := decoder.Decode(&obj); err != nil {
			httpLog.Error(err, "Wrong JSON data received", "url", url, "type", "POST")
			return response.StatusCode, fmt.Errorf("wrong json data received: %s", textData)
		}
		return response.StatusCode, nil
	}
	return 0, err
}
