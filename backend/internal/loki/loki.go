package loki

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type LokiHttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type LokiClient struct {
	baseUrl string
	client  LokiHttpClient
}

func New(baseUrl string, client LokiHttpClient) *LokiClient {
	return &LokiClient{baseUrl: baseUrl, client: client}
}

type Direction string

var (
	FORWARD  Direction = "FORWARD"
	BACKWARD Direction = "BACKWARD"
)

func (d Direction) String() string {
	return string(d)
}

type QueryRangeRequest struct {
	Query     string
	Limit     *int64
	Start     *time.Time
	End       *time.Time
	Direction *Direction
}

func (c *LokiClient) QueryRange(
	ctx context.Context,
	request *QueryRangeRequest,
	logger *slog.Logger,
) (*QueryResponse, error) {
	baseUrl := fmt.Sprintf("%s/loki/api/v1/query_range", c.baseUrl)

	params := url.Values{}
	params.Add("query", request.Query)
	if request.Limit != nil {
		params.Add("limit", fmt.Sprintf("%d", *request.Limit))
	}
	if request.Start != nil {
		params.Add("start", fmt.Sprintf("%d", request.Start.UnixNano()))
	}
	if request.End != nil {
		params.Add("end", fmt.Sprintf("%d", request.End.UnixNano()))
	}
	if request.Direction != nil {
		params.Add("direction", request.Direction.String())
	}

	fullUrl := fmt.Sprintf("%s?%s", baseUrl, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullUrl, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("could not create query_range request: %w", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not receive query_range response: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read query_range body: %w", err)
	}

	if res.StatusCode > 399 {
		logger.Error(fmt.Sprintf("received non 200 status code: %d when querying loki for logs", res.StatusCode), "body", string(body))
		return nil, fmt.Errorf("received non 200 status code for loki query_range: %d", res.StatusCode)
	}

	var typedResp QueryResponse
	err = json.Unmarshal(body, &typedResp)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal query_range response body: %w", err)
	}
	return &typedResp, nil
}

func GetStreamsFromResponseData(data *QueryResponseData) (Streams, error) {
	if data == nil {
		return nil, errors.New("QueryResponseData was nil")
	}
	if data.ResultType != ResultTypeStream {
		return nil, fmt.Errorf("result type was not stream: %s", data.ResultType)
	}
	streams, ok := data.Result.(Streams)
	if !ok {
		return nil, fmt.Errorf("result data type was not Streams, got: %T", data.Result)
	}
	return streams, nil
}

func GetEntriesFromStreams(streams Streams) []*LabeledEntry {
	entries := []*LabeledEntry{}
	for _, stream := range streams {
		for _, entry := range stream.Entries {
			entries = append(entries, &LabeledEntry{Entry: entry, Labels: getFilteredLabels(stream.Labels, allowedLabels)})
		}
	}
	return entries
}

var allowedLabels = []string{"ActivityType", "Name", "Schema", "Table", "Attempt", "metadata_Schema", "metadata_Table"}

func getFilteredLabels(labels LabelSet, keepLabels []string) LabelSet {
	filteredLabels := LabelSet{}
	for _, label := range keepLabels {
		if value, ok := labels[label]; ok {
			filteredLabels[label] = value
		}
	}
	return filteredLabels
}
