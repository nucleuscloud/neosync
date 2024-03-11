package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func main() {
	client, err := promapi.NewClient(promapi.Config{
		Address: "http://localhost:9090",
		// Address: "http://localhost:9091",
	})
	if err != nil {
		panic(err)
	}
	api := promv1.NewAPI(client)
	// date := time.Date(2024, 3, 8, 12, 05, 18, 00, time.Local)
	// value, warnings, err := api.Query(context.Background(), `input_received_total{neosyncJobId="41606319-5eee-4667-8dfd-b8b954b65bf9"}`, date)
	// if err != nil {
	// 	panic(err)
	// }
	jobId := "e4208a7c-bd84-43aa-82d2-7c1503700791"
	// jobId := "41606319-5eee-4667-8dfd-b8b954b65bf9" // staging french
	start := time.Now().Add(-4 * time.Minute)
	// start = time.Date(2024, 3, 8, 18, 0, 0, 0, time.Local)
	query := fmt.Sprintf("input_received_total{neosyncJobId=%q}", jobId)
	value, warnings, err := api.QueryRange(context.Background(), query, promv1.Range{
		// Start: time.Date(2024, 3, 8, 11, 00, 00, 00, time.Local),
		Start: start,
		End:   time.Now(),
		// Step:  24 * time.Hour,
		Step: 1 * time.Minute,
	})
	if err != nil {
		panic(err)
	}
	for _, warning := range warnings {
		fmt.Printf("warning: %v\n", warning)
	}
	switch value.Type() {
	case model.ValMatrix:
		val, ok := value.(model.Matrix)
		if !ok {
			panic("not matrix")
		}
		usage := map[string]int64{}
		for _, stream := range val {
			fmt.Println("METRIC", stream.Metric)
			usage[stream.Metric.String()] = 0
			// for k, v := range stream.Metric {
			// 	fmt.Println("metric", k, v)
			// }
			// get latest, max value
			var latest int64
			for _, value := range stream.Values {
				fmt.Println("matrix value", value.Value.String(), "ts", value.Timestamp.Time().Format(time.RFC3339))

				//
				converted, err := strconv.ParseInt(value.Value.String(), 10, 64)
				if err != nil {
					panic(err)
				}
				if converted > latest {
					latest = converted
				}
			}
			usage[stream.Metric.String()] = latest
		}
		bits, err := json.MarshalIndent(usage, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println("USAGE:", string(bits))

		fmt.Println("found", len(usage), "total metrics for calculation")

		var total int64
		for _, val := range usage {
			total += val
		}
		fmt.Println("total input rows processed:", total)
		// fmt.Println("matrix value", val.String())
	case model.ValScalar:
		val, ok := value.(*model.Scalar)
		if !ok {
			panic("not scaler")
		}
		fmt.Println("scalar value", val.Value.String(), "ts", val.Timestamp.String())
	case model.ValString:
		val, ok := value.(*model.String)
		if !ok {
			panic("not string")
		}
		fmt.Println("string value", val.Value, "ts", val.Timestamp.String())
	case model.ValVector:
		val, ok := value.(model.Vector)
		if !ok {
			panic("not vector")
		}
		for _, sample := range val {
			fmt.Println("-------")
			fmt.Println(sample.Metric.String())
			fmt.Println("vector sample", sample.Value, sample.Timestamp.Time().Format(time.RFC3339))
			fmt.Println("-------")

		}
	default:
		fmt.Println("default value", value.String(), "type", value.Type())

	}
}
