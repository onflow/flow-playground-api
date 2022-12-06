package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var FooCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "foo_counter",
	Help: "The total number of foo",
})
