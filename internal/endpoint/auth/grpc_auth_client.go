package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	auth_v1 "main/internal/lib/gen/auth"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClientImpl struct {
	api auth_v1.AuthClient
	log *slog.Logger
}

func NewGRPClient(
	log *slog.Logger,
	host string,
	port string,
	timeout time.Duration,
	retriesCount int,
) (*GRPCClientImpl, error) {
	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	cc, err := grpc.NewClient(fmt.Sprintf("%s:%s", host, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		))
	if err != nil {
		return nil, err
	}

	grpcClient := auth_v1.NewAuthClient(cc)
	return &GRPCClientImpl{
		api: grpcClient,
		log: log,
	}, nil
}

func (c *GRPCClientImpl) GetRefresh(ctx context.Context, ui UserInfo) (*Token, error) {
	res, err := c.api.NewRefresh(ctx, &auth_v1.NewRefreshRequest{Username: ui.Username})
	if err != nil {
		return nil, err
	}

	return (*Token)(&res.Token), nil
}

func (c *GRPCClientImpl) GetAccess(ctx context.Context, ui UserInfo) (*Token, error) {
	res, err := c.api.NewAccess(ctx, &auth_v1.NewAccessRequest{Username: ui.Username})
	if err != nil {
		return nil, err
	}

	return (*Token)(&res.Token), nil
}

func (c *GRPCClientImpl) GetPair(ctx context.Context, ui UserInfo) (*TokenPair, error) {
	access, err := c.GetAccess(ctx, ui)
	if err != nil {
		return nil, err
	}

	refresh, err := c.GetRefresh(ctx, ui)
	if err != nil {
		return nil, err
	}

	return &TokenPair{Access: Token(*access), Refresh: Token(*refresh)}, nil
}

func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
