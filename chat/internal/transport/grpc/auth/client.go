package auth

import (
	"chat/pkg/logger"
	"context"
	"time"

	authpb "chat/pkg/api/auth_pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthClient struct {
	cfg Config
	conn *grpc.ClientConn
	client authpb.AuthServiceClient 
}

func New(cfg Config) *AuthClient {
	return &AuthClient{
		cfg: cfg,
		conn: nil,
		client: nil,
	}
}

func (ac *AuthClient) connect(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	con, err := grpc.NewClient(
		ac.cfg.Address(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.GetFromCtx(ctx).ErrorContext(ctx, "failed connection to grpc", err)
		return err
	}

	logger.GetFromCtx(ctx).InfoContext(ctx, "listen auth grpc")

	ac.conn = con
	ac.client = authpb.NewAuthServiceClient(con)
	return nil
}

func (ac *AuthClient) Close() error {
	if ac.conn != nil {
		return ac.conn.Close()
	}
	return nil
}

func (ac *AuthClient) GetId(ctx context.Context, token string) (int, error) {
	if ac.conn == nil {
		if err := ac.connect(ctx); err != nil {
			return -1, err
		}
	}

	resp, err := ac.client.Validate(ctx, &authpb.Token{Token: token})
	if err != nil{
		logger.GetFromCtx(ctx).ErrorContext(ctx, "error in grpc service", err)
		return 0, err
	}
	return int(resp.GetUserId()), nil
}