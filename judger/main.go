package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof" // for pprof
	"os"
	"os/signal"
	"time"

	execpb "github.com/criyle/go-judge/pb"
	demopb "github.com/criyle/go-judger-demo/pb"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	envDemoServerURL = "DEMO_SERVER"
	envExecServerURL = "EXEC_SERVER"
	envRelease       = "RELEASE"
	envToken         = "TOKEN"

	defaultDemoServerURL = "localhost:5081"
	defaultExecServerURL = "localhost:5051"
)

const (
	outputLimit = 64 << 10  // 64k
	memoryLimit = 256 << 20 // 256m
	runDir      = "run"
	pathEnv     = "PATH=/usr/local/bin:/usr/bin:/bin"
	noCase      = 8
)

var env = []string{
	pathEnv,
	"HOME=/tmp",
}

var (
	taskHist = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "judger_task_execute_time_seconds",
		Help:    "Time for whole processed case",
		Buckets: prometheus.ExponentialBuckets(time.Millisecond.Seconds(), 1.4, 30), // 1 ms -> 10s
	}, []string{"status"})

	taskSummry = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "judger_task_execute_time",
		Help:       "Summary for the whole process case time",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"status"})
)

var logger *zap.Logger

func init() {
	prometheus.MustRegister(taskHist, taskSummry)
}

func main() {
	_, release := os.LookupEnv(envRelease)
	var err error
	if release {
		logger, _ = zap.NewProduction()
	} else {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, err = config.Build()
	}
	if err != nil {
		log.Fatalln("init logger", err)
	}

	// collect metrics
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

	token := os.Getenv(envToken)
	execServer := defaultExecServerURL
	if e := os.Getenv(envExecServerURL); e != "" {
		execServer = e
	}
	execClient := createExecClient(execServer, token)

	demoServer := defaultDemoServerURL
	if e := os.Getenv(envDemoServerURL); e != "" {
		demoServer = e
	}
	demoClient := createDemoClient(demoServer, token)

	j := newJudger(execClient, demoClient)
	j.Start()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	log.Println("interrupted")
}

func createExecClient(execServer, token string) execpb.ExecutorClient {
	conn, err := createGRPCConnection(execServer, token)
	if err != nil {
		log.Fatalln("client", err)
	}
	return execpb.NewExecutorClient(conn)
}

func createDemoClient(execServer, token string) demopb.DemoBackendClient {
	conn, err := createGRPCConnection(execServer, token)
	if err != nil {
		log.Fatalln("client", err)
	}
	return demopb.NewDemoBackendClient(conn)
}

func createGRPCConnection(addr, token string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			grpc_prometheus.UnaryClientInterceptor,
			grpc_zap.UnaryClientInterceptor(logger),
		)),
		grpc.WithStreamInterceptor(
			grpc_middleware.ChainStreamClient(
				grpc_prometheus.StreamClientInterceptor,
				grpc_zap.StreamClientInterceptor(logger),
			))}
	if token != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(newTokenAuth(token)))
	}
	return grpc.Dial(addr, opts...)
}

type tokenAuth struct {
	token string
}

func newTokenAuth(token string) credentials.PerRPCCredentials {
	return &tokenAuth{token: token}
}

// Return value is mapped to request headers.
func (t *tokenAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.token,
	}, nil
}

func (*tokenAuth) RequireTransportSecurity() bool {
	return false
}
