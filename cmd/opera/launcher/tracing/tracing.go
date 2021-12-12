package tracing

import (
	"errors"
	"fmt"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-opera/tracing"
)

var (
	EnableFlag = cli.BoolFlag{
		Name:  "tracing",
		Usage: "Enable traces collection and reporting",
	}

	AgentEndpointFlag = cli.StringFlag{
		Name:  "tracing.agent",
		Usage: "Jaeger agent endpoint. Default is localhost:6831",
		Value: "localhost:6831",
	}

	DevelopmentFlag = cli.BoolFlag{
		Name:  "tracing.dev",
		Usage: "Use development Jaeger configuration",
	}

	ErrInvalidEndpoint = errors.New("invalid agent endpoint")
)

func Start(ctx *cli.Context) (func(), error) {
	if !ctx.Bool(EnableFlag.Name) {
		return func() {}, nil
	}

	agentEndpoint := ctx.String(AgentEndpointFlag.Name)
	if agentEndpoint == "" {
		return nil, ErrInvalidEndpoint
	}

	// Default config recommended for production
	cfg := jaegercfg.Configuration{
		ServiceName: "opera",
		Reporter: &jaegercfg.ReporterConfig{
			LocalAgentHostPort: agentEndpoint,
		},
	}

	if ctx.Bool(DevelopmentFlag.Name) {
		// Makes sampler collect and report all traces
		cfg.Sampler = &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		}
	}

	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jaegerlog.StdLogger),
		jaegercfg.Metrics(metrics.NullFactory),
	)
	if err != nil {
		return nil, fmt.Errorf("new tracer: %w", err)
	}
	stop := func() {
		closer.Close()
	}

	opentracing.SetGlobalTracer(tracer)
	tracing.SetEnabled(true)

	return stop, nil
}
