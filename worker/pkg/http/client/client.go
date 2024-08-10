package http_client

import (
	"fmt"
	"net/http"
)

// Returns an http client that includes an auth header if the token is not empty or nil
func NewWithAuth(token *string) *http.Client {
	if token == nil || *token == "" {
		return http.DefaultClient
	}
	return NewWithHeaders(GetAuthHeaders(token))
}

func WithAuth(client *http.Client, token *string) *http.Client {
	return WithHeaders(client, GetAuthHeaders(token))
}

// Returns a new http client that will send headers along with the request
func NewWithHeaders(
	headers map[string]string,
) *http.Client {
	return WithHeaders(&http.Client{}, headers)
}

func WithHeaders(
	client *http.Client,
	headers map[string]string,
) *http.Client {
	if client.Transport == nil {
		client.Transport = http.DefaultTransport
	}
	client.Transport = &headerTransport{
		Transport: client.Transport,
		Headers:   headers,
	}
	return client
}

type headerTransport struct {
	Transport http.RoundTripper
	Headers   map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header == nil {
		req.Header = http.Header{}
	}
	for key, value := range t.Headers {
		req.Header.Add(key, value)
	}
	return t.Transport.RoundTrip(req)
}

func GetAuthHeaders(token *string) map[string]string {
	if token == nil || *token == "" {
		return map[string]string{}
	}
	return map[string]string{"Authorization": fmt.Sprintf("Bearer %s", *token)}
}

func MergeMaps(maps ...map[string]string) map[string]string {
	output := map[string]string{}

	for _, input := range maps {
		for k, v := range input {
			output[k] = v
		}
	}

	return output
}
