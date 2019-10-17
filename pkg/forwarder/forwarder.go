package forwarder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/go-playground/validator.v9"
	"io/ioutil"
	"net/http"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func SearchForUpstream(k8sResource string, body *[]byte) ([]UpstreamRequest, error) {
	var upstreamRequests []UpstreamRequest
	var allUpstreams []Upstream
	if err := viper.UnmarshalKey("upstreams", &allUpstreams); err != nil {
		logrus.Warn("Broken config, unable  unmarshal upstreams configurations")
		return nil, fmt.Errorf("Broken config, unable  unmarshal upstreams configurations")
	}
	for _, u := range allUpstreams {
		if u.Resource == k8sResource {
			logrus.Infof("found upstream match, %v, %v", u.Resource, u.Url)
			upstreamRequests = append(upstreamRequests, UpstreamRequest{Upstream: u, Payload: body})
		}
	}
	return upstreamRequests, nil
}

func ForwardValidationRequest(doneChan chan UpstreamResponse, upstreamsRequests *[]UpstreamRequest) error {
	// Upstream Responses slice
	var upstreamResponses []UpstreamResponse
	// Channel for upstream responses results
	urChan := make(chan UpstreamResponse)
	// Run and execute all upstream request in parallel
	for _, ur := range *upstreamsRequests {
		go ur.execUpstreamRequest(urChan, doneChan)
	}
	// Wait for all request to be finished
	for upstreamRes := range urChan {
		// One of the Upstream validation requests is no allowed
		// Stop the execution, send not allowed response to K8S
		// And break the chanel read loop
		if *upstreamRes.IsAllowed == false {
			doneChan <- upstreamRes
			break
		}
		// All good, append result for further logging
		upstreamResponses = append(upstreamResponses, upstreamRes)
		if len(upstreamResponses) == len(*upstreamsRequests) {
			break // All good, all requests has been finished, break the loop
		}
	}
	logrus.Infof("Send to chan is done, waiting to close channel")
	defer close(urChan)
	return nil
}

func (ur UpstreamRequest) execUpstreamRequest(urChan chan UpstreamResponse, doneChan chan UpstreamResponse) {
	var uRes UpstreamResponse
	// As for now, the default behaviour on failure is allow
	isAllowed := true
	logrus.Infof("Forwarding to upstream url: %v, resource: %v", ur.Upstream.Url, ur.Upstream.Resource)
	client := &http.Client{}
	req, err := http.NewRequest("POST", ur.Upstream.Url, bytes.NewBuffer(*ur.Payload))
	if err != nil {
		logrus.Errorf("%v", err)
		// Failed to create new HTTP request, send failure response to channel
		doneChan <- UpstreamResponse{
			IsAllowed:   &isAllowed,
			Message:     fmt.Sprintf("%v", err),
			UpstreamUrl: ur.Upstream.Url}
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	logrus.Infof("Got response for: %v, body: %v", ur.Upstream.Url, resp)
	if err != nil {
		// Failed execute HTTP request, send failure response to channel
		logrus.Errorf("failed execute request POST request, err: %v", err)
		doneChan <- UpstreamResponse{
			IsAllowed:   &isAllowed,
			Message:     fmt.Sprintf("%v", err),
			UpstreamUrl: ur.Upstream.Url}
		return
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(bodyText, &uRes); err != nil {
		// Failed unmarshal HTTP response, send failure response to channel
		logrus.Errorf("Unmarshal upstream response failed, err %v, response body: %s", err, bodyText)
		doneChan <- UpstreamResponse{
			IsAllowed:   &isAllowed,
			Message:     fmt.Sprintf("%v", err),
			UpstreamUrl: ur.Upstream.Url}
		return
	}
	if err := validate.Struct(uRes); err != nil {
		// Failed to validate HTTP response, send failure response to channel
		logrus.Errorf("Invalid response body: %s, %v ", bodyText, err)
		doneChan <- UpstreamResponse{
			IsAllowed:   &isAllowed,
			Message:     fmt.Sprintf("%v", err),
			UpstreamUrl: ur.Upstream.Url}
		return
	}
	uRes.UpstreamUrl = ur.Upstream.Url
	logrus.Infof("Upstream response: [IsAllowed: %v, Message: %s, Url: %s]",
		*uRes.IsAllowed, uRes.Message, uRes.UpstreamUrl)
	// All good, return original upstream response
	urChan <- uRes
}
