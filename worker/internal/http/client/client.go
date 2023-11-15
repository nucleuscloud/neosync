package http_client

import "net/http"

// Returns a new http client that will send headers along with the request
func NewWithHeaders(
	headers map[string]string,
) *http.Client {
	client := &http.Client{
		Transport: &headerTransport{
			Transport: http.DefaultTransport,
			Headers:   headers,
		},
	}
	return client
}

type headerTransport struct {
	Transport http.RoundTripper
	Headers   map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, value := range t.Headers {
		req.Header.Add(key, value)
	}
	return t.Transport.RoundTrip(req)
}
