package main

import (
	"flag"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"log"
	"net/http"
	"time"
)

const (
	serviceName   = "hello-server"
	hostPort      = "0.0.0.0:0"
	debug         = false
	sameSpan      = false
	traceID128Bit = true
)

var collectorHost = flag.String("collector_host", "localhost", "Host for Jaeger Collector")
var collectorPort = flag.String("collector_port", "6831", "Port for Jaeger Collector")

func handler(w http.ResponseWriter, r *http.Request) {
	wireContext, _ := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))
	span := opentracing.StartSpan(
		"/",
		opentracing.ChildOf(wireContext))
	defer span.Finish()
	tm := time.Now().Format(time.RFC1123)
	w.Header().Set("x_service", "handler1")
	w.Write([]byte("The time is " + tm))
}

func handler2(w http.ResponseWriter, r *http.Request) {
	wireContext, _ := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))
	span := opentracing.StartSpan(
		"/",
		opentracing.ChildOf(wireContext))
	defer span.Finish()
	tm := time.Now().Format(time.RFC1123)
	w.Header().Set("x_service", "handler2")
	w.Write([]byte("The time is " + tm))
}

func main() {
	flag.Parse()
	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LocalAgentHostPort:  *collectorHost + ":" + *collectorPort,
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
		},
	}
	closer, err := cfg.InitGlobalTracer(
		"backend",
		config.Logger(jaeger.StdLogger),
	)

	if err != nil {
		log.Printf("Could not initialize jaeger tracer: %s", err.Error())
		return
	}

	defer closer.Close()

	http.HandleFunc("/", handler)
	http.HandleFunc("/2", handler2)
	http.ListenAndServe(":9001", nil)
}
