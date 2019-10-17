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

func ForwardValidationRequests(doneChan chan UpstreamResponse, upstreamsRequests *[]UpstreamRequest) {
	// Upstream Responses slice
	var upstreamResponses []UpstreamResponse
	// Channel for upstream responses results
	urChan := make(chan UpstreamResponse)
	// Upstream response channel open flag
	urChanOpen := true
	// Run and execute all upstream request in parallel
	for _, ur := range *upstreamsRequests {
		go ur.execUpstreamRequest(urChan, &urChanOpen)
	}
	// Wait for all request to be finished
	for upstreamRes := range urChan {
		// One of the Upstream validation requests is no allowed
		// Stop the execution, send not allowed response to K8S
		// And break the chanel read loop
		if *upstreamRes.IsAllowed == false {
			doneChan <- upstreamRes // Send admission response to K8S
			urChanOpen = false      // Close the upstream results channel
			break
		}
		// All good, append result for further logging
		upstreamResponses = append(upstreamResponses, upstreamRes)
		if len(upstreamResponses) == len(*upstreamsRequests) {
			doneChan <- upstreamRes // Send admission response to K8S
			break // All good, all requests has been finished, break the loop
		}
	}
	logrus.Infof("Upstream request triggered, waiting for all request finish")
	defer close(urChan)
}

func (ur UpstreamRequest) execUpstreamRequest(urChan chan UpstreamResponse, urChanOpen *bool) {
	var uRes UpstreamResponse
	// As for now, the default behaviour on system (not upstream response) failure is allow
	isAllowed := true
	logrus.Infof("Forwarding to upstream url: %v, resource: %v", ur.Upstream.Url, ur.Upstream.Resource)
	client := &http.Client{}
	req, err := http.NewRequest("POST", ur.Upstream.Url, bytes.NewBuffer(*ur.Payload))
	if err != nil {
		logrus.Errorf("%v", err)
		// Failed to create new HTTP request, send failure response to channel
		urChan <- UpstreamResponse{
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
		urChan <- UpstreamResponse{
			IsAllowed:   &isAllowed,
			Message:     fmt.Sprintf("%v", err),
			UpstreamUrl: ur.Upstream.Url}
		return
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(bodyText, &uRes); err != nil {
		// Failed unmarshal HTTP response, send failure response to channel
		logrus.Errorf("Unmarshal upstream response failed, err %v, response body: %s", err, bodyText)
		urChan <- UpstreamResponse{
			IsAllowed:   &isAllowed,
			Message:     fmt.Sprintf("%v", err),
			UpstreamUrl: ur.Upstream.Url}
		return
	}
	if err := validate.Struct(uRes); err != nil {
		// Failed to validate HTTP response, send failure response to channel
		logrus.Errorf("Invalid response body: %s, %v ", bodyText, err)
		urChan <- UpstreamResponse{
			IsAllowed:   &isAllowed,
			Message:     fmt.Sprintf("%v", err),
			UpstreamUrl: ur.Upstream.Url}
		return
	}
	uRes.UpstreamUrl = ur.Upstream.Url
	logrus.Infof("Upstream response: [IsAllowed: %v, Message: %s, Url: %s]",
		*uRes.IsAllowed, uRes.Message, uRes.UpstreamUrl)
	// All good, return original upstream response
	logrus.Infof("Is Upstream response results channel open: %v", *urChanOpen)
	if *urChanOpen == true {
		urChan <- uRes
	}

}
