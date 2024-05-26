package httpclient

import (
	"io"
	"net/http"

	"github.com/minhvuongrbs/webhook-service/pkg/logging"
)

type roundTripper struct{}

func NewRoundTripper() roundTripper {
	return roundTripper{}
}

func (t roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	l := logging.FromContext(req.Context()).
		Named("HttpClient")

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		l.Warnf("failed to read request body: %s", err)
	}

	l.Infof("send request, url=%v, body=%v", req.URL, reqBody)
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return resp, err
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Warnf("failed to read response body: %s", err)
	}
	l.Infof(
		"receive response, url=%v, body=%v, status_code=%d", resp.Request.URL, respBody, resp.StatusCode)

	return resp, err
}
