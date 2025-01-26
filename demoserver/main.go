package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof" // for pprof
	"os"
	"os/signal"

	execpb "github.com/criyle/go-judge/pb"
	"github.com/criyle/go-judger-demo/pb"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

const (
	envGRPCAddr   = "GRPC_ADDR"
	envExecServer = "EXEC_SERVER"
	envToken      = "TOKEN"
	envRelease    = "RELEASE"
	envMongoURI   = "MONGODB_URI"
)

func InterceptorLogger(l *zap.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2)

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			value := fields[i+1]

			switch v := value.(type) {
			case string:
				f = append(f, zap.String(key.(string), v))
			case int:
				f = append(f, zap.Int(key.(string), v))
			case bool:
				f = append(f, zap.Bool(key.(string), v))
			default:
				f = append(f, zap.Any(key.(string), v))
			}
		}

		logger := l.WithOptions(zap.AddCallerSkip(1)).With(f...)

		switch lvl {
		case logging.LevelDebug:
			logger.Debug(msg)
		case logging.LevelInfo:
			logger.Info(msg)
		case logging.LevelWarn:
			logger.Warn(msg)
		case logging.LevelError:
			logger.Error(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

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
	prom := grpc_prometheus.NewServerMetrics(grpc_prometheus.WithServerHandlingTimeHistogram())
	prometheus.MustRegister(prom)
	grpclog.SetLoggerV2(zapgrpc.NewLogger(logger))
	streamMiddleware := []grpc.StreamServerInterceptor{
		prom.StreamServerInterceptor(),
		logging.StreamServerInterceptor(InterceptorLogger(logger)),
		grpc_recovery.StreamServerInterceptor(),
	}
	unaryMiddleware := []grpc.UnaryServerInterceptor{
		prom.UnaryServerInterceptor(),
		logging.UnaryServerInterceptor(InterceptorLogger(logger)),
		grpc_recovery.UnaryServerInterceptor(),
	}
	if token != "" {
		authFunc := grpcTokenAuth(token)
		streamMiddleware = append(streamMiddleware, grpc_auth.StreamServerInterceptor(authFunc))
		unaryMiddleware = append(unaryMiddleware, grpc_auth.UnaryServerInterceptor(authFunc))
	}
	grpcServer = grpc.NewServer(
		grpc.ChainStreamInterceptor(streamMiddleware...),
		grpc.ChainUnaryInterceptor(unaryMiddleware...),
	)
	pb.RegisterDemoBackendServer(grpcServer, ds)

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
	prom := grpc_prometheus.NewClientMetrics(grpc_prometheus.WithClientHandlingTimeHistogram())
	prometheus.MustRegister(prom)
	grpclog.SetLoggerV2(zapgrpc.NewLogger(logger))
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			prom.UnaryClientInterceptor(),
			logging.UnaryClientInterceptor(InterceptorLogger(logger)),
		),
		grpc.WithChainStreamInterceptor(
			prom.StreamClientInterceptor(),
			logging.StreamClientInterceptor(InterceptorLogger(logger)),
		)}
	if token != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(newTokenAuth(token)))
	}
	return grpc.NewClient(addr, opts...)
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
