/*
 * Flow Playground
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package telemetry

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	existStatusFailure = "failure"
	exitStatusSuccess  = "success"
)

var (
	registered               bool
	requestStartedCounter    prometheus.Counter
	requestCompletedCounter  prometheus.Counter
	resolverStartedCounter   *prometheus.CounterVec
	resolverCompletedCounter *prometheus.CounterVec
	timeToResolveField       *prometheus.HistogramVec
	timeToHandleRequest      *prometheus.HistogramVec
	staleProjectGauge        prometheus.Gauge
	totalProjectGauge        prometheus.Gauge
	ServerErrorCounter       prometheus.Counter
	UserErrorCounter         prometheus.Counter
)

type (
	RequestsMetrics struct{}
)

var _ interface {
	graphql.HandlerExtension
	graphql.OperationInterceptor
	graphql.ResponseInterceptor
	graphql.FieldInterceptor
} = RequestsMetrics{}

func NewMetrics() graphql.HandlerExtension {
	RegisterMetrics()
	return RequestsMetrics{}
}

func RegisterMetrics() {
	if !registered {
		RegisterOn(prometheus.DefaultRegisterer)
		err := registerProjectJobs()
		if err != nil {
			log.Printf("Failed to register job for stale project metrics: %s", err.Error())
		}
		registered = true
	}
}

func RegisterOn(registerer prometheus.Registerer) {
	requestStartedCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "graphql_request_started_total",
			Help: "Total number of requests started on the graphql server.",
		},
	)

	requestCompletedCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "graphql_request_completed_total",
			Help: "Total number of requests completed on the graphql server.",
		},
	)

	resolverStartedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "graphql_resolver_started_total",
			Help: "Total number of resolver started on the graphql server.",
		},
		[]string{"object", "field"},
	)

	resolverCompletedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "graphql_resolver_completed_total",
			Help: "Total number of resolver completed on the graphql server.",
		},
		[]string{"object", "field"},
	)

	timeToResolveField = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "graphql_resolver_duration_ms",
		Help:    "The time taken to resolve a field by graphql server.",
		Buckets: prometheus.ExponentialBuckets(1, 2, 11),
	}, []string{"exitStatus", "object", "field"})

	timeToHandleRequest = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "graphql_request_duration_ms",
		Help:    "The time taken to handle a request by graphql server.",
		Buckets: prometheus.ExponentialBuckets(1, 2, 11),
	}, []string{"exitStatus"})

	staleProjectGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "stale_projects_total",
		Help: fmt.Sprintf("The total number of projects not accessed within the last %s days.",
			strconv.FormatFloat(staleDuration.Hours()/24, 'f', -1, 64)),
	})

	totalProjectGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "total_projects",
		Help: "The total number of projects in the database.",
	})

	ServerErrorCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "server_error_total",
			Help: "Total number of internal server errors.",
		},
	)

	UserErrorCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "user_error_total",
			Help: "Total number of bad requests from users.",
		},
	)

	registerer.MustRegister(
		requestStartedCounter,
		requestCompletedCounter,
		resolverStartedCounter,
		resolverCompletedCounter,
		timeToResolveField,
		timeToHandleRequest,
		staleProjectGauge,
		totalProjectGauge,
		ServerErrorCounter,
		UserErrorCounter,
	)
}

func UnRegisterMetrics() {
	if registered {
		UnRegisterFrom(prometheus.DefaultRegisterer)
		registered = false
	}
}

func UnRegisterFrom(registerer prometheus.Registerer) {
	registerer.Unregister(requestStartedCounter)
	registerer.Unregister(requestCompletedCounter)
	registerer.Unregister(resolverStartedCounter)
	registerer.Unregister(resolverCompletedCounter)
	registerer.Unregister(timeToResolveField)
	registerer.Unregister(timeToHandleRequest)
	registerer.Unregister(staleProjectGauge)
	registerer.Unregister(totalProjectGauge)
	registerer.Unregister(ServerErrorCounter)
	registerer.Unregister(UserErrorCounter)
}

func (a RequestsMetrics) ExtensionName() string {
	return "Prometheus"
}

func (a RequestsMetrics) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

func (a RequestsMetrics) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	requestStartedCounter.Inc()
	return next(ctx)
}

func (a RequestsMetrics) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	errList := graphql.GetErrors(ctx)

	var exitStatus string
	if len(errList) > 0 {
		exitStatus = existStatusFailure
	} else {
		exitStatus = exitStatusSuccess
	}

	oc := graphql.GetOperationContext(ctx)
	observerStart := oc.Stats.OperationStart

	timeToHandleRequest.With(prometheus.Labels{"exitStatus": exitStatus}).
		Observe(float64(time.Since(observerStart).Nanoseconds() / int64(time.Millisecond)))

	requestCompletedCounter.Inc()

	return next(ctx)
}

func (a RequestsMetrics) InterceptField(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	fc := graphql.GetFieldContext(ctx)

	resolverStartedCounter.WithLabelValues(fc.Object, fc.Field.Name).Inc()

	observerStart := time.Now()

	res, err := next(ctx)

	var exitStatus string
	if err != nil {
		exitStatus = existStatusFailure
	} else {
		exitStatus = exitStatusSuccess
	}

	timeToResolveField.WithLabelValues(exitStatus, fc.Object, fc.Field.Name).
		Observe(float64(time.Since(observerStart).Nanoseconds() / int64(time.Millisecond)))

	resolverCompletedCounter.WithLabelValues(fc.Object, fc.Field.Name).Inc()

	return res, err
}
