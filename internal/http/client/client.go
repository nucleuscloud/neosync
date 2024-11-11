package http_client

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

// Returns an http client that includes an auth header if the token is not empty or nil
func NewWithBearerAuth(token *string) *http.Client {
	if token == nil || *token == "" {
		return http.DefaultClient
	}
	return NewWithHeaders(GetBearerAuthHeaders(token))
}

// Creates an Authorization Bearer Auth header on the http client
func WithBearerAuth(client *http.Client, token *string) *http.Client {
	return WithHeaders(client, GetBearerAuthHeaders(token))
}

// Creates an Authorization Basic Auth header on the http Client
func WithBasicAuth(client *http.Client, username, password string) *http.Client {
	return WithHeaders(client, GetBasicAuthHeaders(username, password))
}

// Treats value as the opaque value for the Authorization header
func WithAuth(client *http.Client, value string) *http.Client {
	return WithHeaders(client, GetAuthHeaders(value))
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
		req.Header.Set(key, value)
	}
	return t.Transport.RoundTrip(req)
}

func GetBearerAuthHeaders(token *string) map[string]string {
	if token == nil || *token == "" {
		return map[string]string{}
	}
	return GetAuthHeaders(fmt.Sprintf("Bearer %s", *token))
}

func GetBasicAuthHeaders(username, password string) map[string]string {
	return GetAuthHeaders(fmt.Sprintf("Basic %s", getBasicAuthValue(username, password)))
}

func GetAuthHeaders(value string) map[string]string {
	return map[string]string{"Authorization": value}
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

func getBasicAuthValue(username, password string) string {
	auth := fmt.Sprintf("%s:%s", username, password)
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
