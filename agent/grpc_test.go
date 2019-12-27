// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package agent

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/doublemo/balala/agent/transport"
	"github.com/doublemo/balala/cores/proto/pb"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd/etcdv3"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	stdzipkin "github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
)

func TestGRPC(t *testing.T) {
	logger := log.NewLogfmtLogger(os.Stderr)
	client, err := etcdv3.NewClient(context.Background(), []string{"127.0.0.1:2379"}, etcdv3.ClientOptions{
		CACert:        "",
		Cert:          "",
		Key:           "",
		Username:      "",
		Password:      "",
		DialTimeout:   3 * time.Second,
		DialKeepAlive: 3 * time.Second,
	})

	if err != nil {
		t.Fatal(err)
		return
	}

	instancer, err := transport.MakeInstancer(client, "/services/balala/1", logger)
	if err != nil {
		t.Fatal(err)
		return
	}

	//tracer := stdopentracing.GlobalTracer() // no-op
	//zipkinTracer, _ := stdzipkin.NewTracer(nil, stdzipkin.WithNoopTracer(true))

	reporter := zipkinhttp.NewReporter("http://192.168.31.20:9411/api/v2/spans")
	defer reporter.Close()

	zEP, _ := stdzipkin.NewEndpoint("agent", "9094")
	tracer := stdopentracing.GlobalTracer() // no-op

	//zipkinTracer, _ := stdzipkin.NewTracer(nil, stdzipkin.WithNoopTracer(true))
	zipkinTracer, err := stdzipkin.NewTracer(reporter, stdzipkin.WithLocalEndpoint(zEP))
	tracer = zipkinot.Wrap(zipkinTracer)
	factory := transport.MakeFactoryCall(logger, tracer, zipkinTracer, []byte("balala"))
	fn := transport.MakeRetry(instancer, factory, 1, 500*time.Millisecond, logger)
	t.Log(fn)

	ret, err := fn(context.Background(), &pb.Request{
		Body:    make([]byte, 1),
		Command: 11,
	})

	t.Log(ret, err)
}
