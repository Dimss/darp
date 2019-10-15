package forwarder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"gopkg.in/go-playground/validator.v8"
)

var validate *validator.Validate

func init() {
	config := &validator.Config{TagName: "validate"}
	validate = validator.New(config)
}
func SearchUpstream(resource string) *Upstream {
	var upstreams []Upstream
	if err := viper.UnmarshalKey("upstreams", &upstreams); err != nil {
		logrus.Warn("Broken config, unable  unmarshal upstreams configurations")
		return nil
	}
	for _, u := range upstreams {
		if u.Resource == resource {
			logrus.Infof("found upstream match, %v, %v", u.Resource, u.Url)
			return &Upstream{Url: u.Url, Resource: u.Resource}
		}
	}
	return nil
}

func (u Upstream) ForwardValidationRequest() (*UpstreamResponse, error) {
	logrus.Infof("Forwarding to upstream url: %v, resource: %v", u.Url, u.Resource)
	client := &http.Client{}
	req, err := http.NewRequest("POST", u.Url, bytes.NewBuffer(u.Body))
	if err != nil {
		logrus.Errorf("%v", err)
		return nil, fmt.Errorf("failed compose new POST request for url: %s", u.Url)
	}
	req.Header.Set("Content-Type", "application/json")
	// Exec request
	resp, err := client.Do(req)
	logrus.Infof("Upstream: %v response: %v", u.Url, resp)
	if err != nil {
		logrus.Errorf("failed execute request POST request, err: %v", err)
		return nil, fmt.Errorf("failed execute POST request for url: %s", u.Url)
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	var upstreamResponse UpstreamResponse
	if err := json.Unmarshal(bodyText, &upstreamResponse); err != nil {
		logrus.Errorf("Unmarshal upstream response failed, err %v, response body: %s", err, bodyText)
		return nil, fmt.Errorf("bad upstream response, url: %v", u.Url)
	}

	if err := validate.Struct(upstreamResponse); err != nil {
		logrus.Errorf("Invalid response body: %s, %v ", bodyText, err)
		return nil, fmt.Errorf("bad upstream response, url: %v", u.Url)
	}

	logrus.Infof("Upstream response: %v", upstreamResponse)
	return &upstreamResponse, nil
}
