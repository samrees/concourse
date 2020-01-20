package metrics

import (
	"fmt"
	"net/http"

	"github.com/concourse/flag"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/metric"
	"go.opentelemetry.io/otel/exporter/metric/prometheus"
)

type Prometheus struct {
	BindIP   flag.IP `long:"prometheus-bind-ip"   default:"0.0.0.0" description:"IP address on which to listen for web traffic."`
	BindPort uint16  `long:"prometheus-bind-port" default:"8080"    description:"Port on which to listen for HTTP traffic."`
}

var (
	Configured bool

	noopMeter = metric.NoopMeter{}
)

func Counter(name string) *metric.Int64Counter {
	if !Configured {
		counter := noopMeter.NewInt64Counter(name)
		return &counter
	}

	counter := global.MeterProvider().Meter("concourse").NewInt64Counter(name)
	return &counter
}

func (p Prometheus) Configure() (controller metric.Provider, err error) {
	controller, hf, err := prometheus.InstallNewPipeline(prometheus.Config{})
	if err != nil {
		err = fmt.Errorf("prometheus install new pipeline: %w")
		return
	}

	http.HandleFunc("/", hf)
	go func() {
		_ = http.ListenAndServe(
			fmt.Sprintf("%s:%d", p.BindIP.String(), p.BindPort),
			nil,
		)
	}()

	return
}
