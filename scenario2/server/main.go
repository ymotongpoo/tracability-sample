// Copyright 2021 Yoshi Yamaguchi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"io"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/compute/metadata"
	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/label"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	apitrace "go.opentelemetry.io/otel/trace"
)

func initTracer() func() {
	projectID, err := metadata.ProjectID()
	if err != nil {
		log.Fatalf("metadata.ProjectID: %v", err)
	}
	_, flush, err := texporter.InstallNewPipeline(
		[]texporter.Option{
			texporter.WithProjectID(projectID),
		},
		sdktrace.WithConfig(sdktrace.Config{
			DefaultSampler: sdktrace.AlwaysSample(),
		}),
	)
	if err != nil {
		log.Fatalf("texporter.InstallNewPipeline: %v", err)
	}
	return flush
}

func main() {
	handler := otelhttp.NewHandler(http.HandlerFunc(barHandler), "root")
	http.Handle("/", handler)
	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		log.Fatalf("error on running HTTP server: %v", err)
	}
}

func barHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := apitrace.SpanFromContext(ctx)
	ik := label.Key("instance")
	instance := baggage.Value(ctx, ik)
	span.AddEvent("handling in", apitrace.WithAttributes(ik.String(instance.AsString())))
	time.Sleep(200 * time.Millisecond) // Simulate blocking call
	io.WriteString(w, "OK")
}
