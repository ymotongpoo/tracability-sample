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
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptrace"
	"time"

	"cloud.google.com/go/compute/metadata"
	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/label"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
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

var BarEndpoint string

func main() {
	log.Printf("target endpoint: %v", BarEndpoint)

	flush := initTracer()
	defer flush()
	tracer := otel.GetTracerProvider().Tracer("tracability/senario2")
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	t := time.NewTicker(15 * time.Second)
	for range t.C {
		func() {
			ctx := baggage.ContextWithValues(context.Background(),
				label.String("username", "sample-user"),
			)
			ctx, span := tracer.Start(ctx, "main",
				trace.WithAttributes(semconv.PeerServiceKey.String("senario2-main")))
			defer span.End()
			ctx = httptrace.WithClientTrace(ctx, otelhttptrace.NewClientTrace(ctx))
			Foo(ctx, client)
		}()
	}
}

func Foo(ctx context.Context, c http.Client) {
	tr := otel.Tracer("foo")
	_, span := tr.Start(ctx, "foo-span",
		trace.WithAttributes(semconv.PeerServiceKey.String("senario2-foo")))
	defer span.End()

	req, _ := http.NewRequestWithContext(ctx, "GET", BarEndpoint, nil)
	res, err := c.Do(req)
	if err != nil {
		log.Printf("error on requesting server: %v", err)
	}
	defer res.Body.Close()
	b, _ := ioutil.ReadAll(res.Body)
	log.Printf("resepose from server: %v", string(b))
	time.Sleep(50 * time.Millisecond)
}
