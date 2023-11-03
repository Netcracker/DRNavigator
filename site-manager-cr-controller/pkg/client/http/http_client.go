package http_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/netcracker/drnavigator/site-manager-cr-controller/logger"
)

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

	log := logger.SimpleLogger()
	log.Debugf("REST url: %s", url)
	log.Debugf("REST request: GET")

	var response *http.Response

	for retry > 0 {
		response, err = client.Do(req)
		if err != nil {
			log.Errorf("Request error: %s", err)
			retry -= 1
		} else {
			break
		}
	}

	if response != nil {
		defer response.Body.Close()
		textData, err := io.ReadAll(response.Body)
		if err != nil {
			log.Errorf("Can't retreive response body: %s", err)
			return 0, err
		}
		log.Debugf("Status: %s", response.Status)
		log.Debugf("Response: \n%s", textData)
		if err := json.Unmarshal(textData, &obj); err != nil {
			log.Errorf("Wrong JSON data received: %s", err)
			return 0, err
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

	log := logger.SimpleLogger()
	log.Debugf("REST url: %s", url)
	log.Debugf("REST body: %s", byteBody)

	var response *http.Response

	for retry > 0 {
		response, err = client.Do(req)
		if err != nil {
			log.Errorf("Request error: %s", err)
			retry -= 1
		} else {
			break
		}
	}

	if response != nil {
		defer response.Body.Close()
		textData, err := io.ReadAll(response.Body)
		if err != nil {
			log.Errorf("Can't retreive response body: %s", err)
			return 0, err
		}
		log.Debugf("Status: %s", response.Status)
		log.Debugf("Response: \n%s", textData)
		if err := json.Unmarshal(textData, &obj); err != nil {
			log.Errorf("Wrong JSON data received: %s", err)
			return 0, err
		}
		return response.StatusCode, nil
	}
	return 0, err
}
