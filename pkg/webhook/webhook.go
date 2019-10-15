package webhook

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"net/http"
	"github.com/darp/pkg/forwarder"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

func ValidateWebHookHandler(w http.ResponseWriter, r *http.Request) {
	logrus.Info("Handling webhook")
	var body []byte
	// Read request body
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	// K8S sends POST request with the admission webhook data,
	// the body can't be empty, but if it is,
	// further processing will be stopped and empty
	// admission response will be sent to K8S API
	if len(body) == 0 {
		errMessage := "The body is empty, can't proceed the request"
		sendAdmissionValidationResponse(w, false, errMessage)
		logrus.Errorf(errMessage)
		return
	}
	ar := v1beta1.AdmissionReview{}
	// Try to decode body into Admission Review object
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		logrus.Errorf("Error during deserializing request body: %v", err)
		sendAdmissionValidationResponse(w, false, "error during deserializing request body")
		return
	}
	upstream := forwarder.SearchUpstream(ar.Request.Resource.Resource)
	if upstream == nil {
		logrus.Infof("No upstreams was configured for resource: %v", ar.Request.Resource.Resource)
		sendAdmissionValidationResponse(w, true, "Automatic allow response")
		return
	}
	logrus.Infof("Upstreams was found for resource: %v, proxying request", ar.Request.Resource.Resource)
	upstream.Body = body
	upstreamResponse, err := upstream.ForwardValidationRequest()
	if err != nil {
		logrus.Errorf("Error during validating forwarding request: %v", err)
		sendAdmissionValidationResponse(w, true, "Automatic allow response")
		return
	}
	sendAdmissionValidationResponse(w, upstreamResponse.IsAllowed, upstreamResponse.Message)

}

func LivenessHandler(w http.ResponseWriter, r *http.Request) {
	var upstreams []forwarder.Upstream
	if err := viper.UnmarshalKey("upstreams", &upstreams); err != nil {
		logrus.Warn("Broken config, unable  unmarshal upstreams configurations")
	}
	logrus.Info(upstreams)
	if _, err := w.Write([]byte("OK")); err != nil {
		logrus.Error("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
		return
	}

	logrus.Info("I'm alive...")
}

func sendAdmissionValidationResponse(w http.ResponseWriter, isAllowed bool, message string) {
	var admissionResponse *v1beta1.AdmissionResponse
	admissionResponse = &v1beta1.AdmissionResponse{Allowed: isAllowed, Result: &metav1.Status{Message: message}}
	admissionReview := v1beta1.AdmissionReview{}
	admissionReview.Response = admissionResponse
	resp, err := json.Marshal(admissionReview)
	if err != nil {
		logrus.Errorf("Error during marshaling admissionResponse object: %v", err)
		http.Error(w, fmt.Sprintf("Error during marshaling admissionResponse object: %w", err), http.StatusInternalServerError)
	}
	logrus.Info("Sending response to API server")
	if _, err := w.Write(resp); err != nil {
		logrus.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}
