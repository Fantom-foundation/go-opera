package tracing

import (
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-lachesis/tracing"
)

var EnableFlag = cli.BoolFlag{
	Name:  "tracing",
	Usage: "Enable traces collection and reporting",
}

func Start(ctx *cli.Context) (stop func()) {
	if !ctx.Bool(EnableFlag.Name) {
		stop = func() {}
		return
	}

	cfg := jaegercfg.Configuration{
		ServiceName: "lachesis",
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst, // to sample every trace
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true, // to log every span via configured Logger
		},
	}

	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		panic(err)
	}
	stop = func() {
		closer.Close()
	}

	opentracing.SetGlobalTracer(tracer)
	tracing.SetEnabled(true)
	return
}
