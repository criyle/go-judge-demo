package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/criyle/go-judge-demo/pb"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
)

const (
	envToken          = "TOKEN"
	envDemoServerAddr = "DEMO_SERVER"
	envRelease        = "RELEASE"
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

	token := os.Getenv(envToken)
	srvAddr := "localhost:5081"
	if e := os.Getenv(envDemoServerAddr); e != "" {
		srvAddr = e
	}

	prom := grpc_prometheus.NewClientMetrics(grpc_prometheus.WithClientHandlingTimeHistogram())
	prometheus.MustRegister(prom)
	grpclog.SetLoggerV2(zapgrpc.NewLogger(logger))
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials()),
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
	conn, err := grpc.NewClient(srvAddr, opts...)
	if err != nil {
		log.Fatalln("client", err)
	}
	client := pb.NewDemoBackendClient(conn)

	r := gin.New()
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))

	r.Use(static.ServeRoot("/", "dist"))
	r.NoRoute(serveIndex)

	apiGroup := r.Group("/api")
	api := &api{client: client}
	api.Register(apiGroup)

	wsGroup := r.Group("/api/ws")
	ju := newJudgeUpdater(client, logger)
	ju.Register(wsGroup)
	sh := &shellHandle{client: client, logger: logger}
	sh.Register(wsGroup)

	addr := ":5000"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	srv := http.Server{
		Handler: r,
		Addr:    addr,
	}
	srv.ListenAndServe()
}

func serveIndex(c *gin.Context) {
	if strings.HasPrefix(c.Request.URL.Path, "/api") {
		c.AbortWithStatus(http.StatusNotImplemented)
		return
	}
	c.File("dist/index.html")
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
