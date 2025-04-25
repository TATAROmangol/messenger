package transport

import (
	pbapi "auth_service/pkg/api"
	"auth_service/pkg/logger"
	"auth_service/pkg/postgres"
	"context"
	"fmt"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

type Service struct {
	pbapi.UnimplementedAuthServiceServer
}

func New() *Service {
	return &Service{}
}

func (x *Service) Validate(ctx context.Context, in *pbapi.Token) (*pbapi.ValidateResponse, error) {
	fmt.Println("token: ", in.GetToken())
	token, err := jwt.ParseWithClaims(in.GetToken(), &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTKey), nil
	})
	exp, _ := token.Claims.GetExpirationTime()
	if exp.Before(time.Now()) {
		return &pbapi.ValidateResponse{UserId: -1}, fmt.Errorf("Token has expired")
	}
	id, err := postgres.GetIdByToken(in.GetToken())
	return &pbapi.ValidateResponse{UserId: int32(id)}, err
}

func (x *Service) GetUser(ctx context.Context, in *pbapi.Token) (*pbapi.GetUserResponse, error) {
	login, email, name, err := postgres.GetRecordByToken(in.GetToken())
	return &pbapi.GetUserResponse{Login: login, Email: email, Name: name}, err
}

func UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	guid := uuid.New().String()
	ctx, _ = logger.New(ctx)
	ctx = context.WithValue(ctx, logger.RequestIDKey, guid)
	hand, err := handler(ctx, req)
	if err != nil {
		logger.Warn(ctx, "Failed to call RPC", zap.Error(err))
	}
	return hand, err
}
