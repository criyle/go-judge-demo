package main

import (
	"context"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof" // for pprof
	"os"
	"os/signal"

	execpb "github.com/criyle/go-judge/pb"
	"github.com/criyle/go-judger-demo/pb"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	envGRPCAddr   = "GRPC_ADDR"
	envExecServer = "EXEC_SERVER"
	envToken      = "TOKEN"
	envRelease    = "RELEASE"
	envMongoURI   = "MONGODB_URI"
)

func main() {
	_, release := os.LookupEnv(envRelease)
	var (
		logger *zap.Logger
		err    error
	)
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

	db := getDB()
	token := os.Getenv(envToken)
	grpcAddr := os.Getenv(envGRPCAddr)
	if grpcAddr == "" {
		grpcAddr = ":5081"
	}
	execServerAddr := os.Getenv(envExecServer)
	if execServerAddr == "" {
		execServerAddr = "localhost:5051"
	}
	execClient := createExecClient(execServerAddr, token, logger)
	ds := newDemoServer(db, execClient, logger)

	var grpcServer *grpc.Server
	grpc_zap.ReplaceGrpcLoggerV2(logger)
	streamMiddleware := []grpc.StreamServerInterceptor{
		grpc_prometheus.StreamServerInterceptor,
		grpc_zap.StreamServerInterceptor(logger),
		grpc_recovery.StreamServerInterceptor(),
	}
	unaryMiddleware := []grpc.UnaryServerInterceptor{
		grpc_prometheus.UnaryServerInterceptor,
		grpc_zap.UnaryServerInterceptor(logger),
		grpc_recovery.UnaryServerInterceptor(),
	}
	if token != "" {
		authFunc := grpcTokenAuth(token)
		streamMiddleware = append(streamMiddleware, grpc_auth.StreamServerInterceptor(authFunc))
		unaryMiddleware = append(unaryMiddleware, grpc_auth.UnaryServerInterceptor(authFunc))
	}
	grpcServer = grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(streamMiddleware...)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaryMiddleware...)),
	)
	pb.RegisterDemoBackendServer(grpcServer, ds)
	grpc_prometheus.Register(grpcServer)
	grpc_prometheus.EnableHandlingTimeHistogram()

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalln(err)
	}
	go func() {
		logger.Sugar().Info("Starting gRPC server at ", grpcAddr)
		logger.Sugar().Info("gRPC serve finished: ", grpcServer.Serve(lis))
	}()

	// collect metrics
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":5082", nil)
	}()

	// Graceful shutdown...
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	grpcServer.GracefulStop()
}

func grpcTokenAuth(token string) func(context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		reqToken, err := grpc_auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err
		}
		if reqToken != token {
			return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
		}
		return ctx, nil
	}
}

func createExecClient(execServer, token string, logger *zap.Logger) execpb.ExecutorClient {
	conn, err := createGRPCConnection(execServer, token, logger)
	if err != nil {
		log.Fatalln("client", err)
	}
	return execpb.NewExecutorClient(conn)
}

func createGRPCConnection(addr, token string, logger *zap.Logger) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
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
