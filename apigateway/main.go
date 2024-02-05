package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/criyle/go-judger-demo/pb"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	envToken          = "TOKEN"
	envDemoServerAddr = "DEMO_SERVER"
	envRelease        = "RELEASE"
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

	token := os.Getenv(envToken)
	srvAddr := "localhost:5081"
	if e := os.Getenv(envDemoServerAddr); e != "" {
		srvAddr = e
	}

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials()),
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
	conn, err := grpc.Dial(srvAddr, opts...)
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
