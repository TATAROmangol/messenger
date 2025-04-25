package httpserver

import (
	"chat/pkg/logger"
	"context"
	"net/http"
)

//go:generate mockgen -destination=./mock/mock_middleware.go -package=mock -source=middleware.go

type Auther interface {
	GetId(ctx context.Context, token string) (int, error)
}

func InitLoggerCtxMiddleware(ctx context.Context) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := logger.InitFromCtx(r.Context(), logger.GetFromCtx(ctx))
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func AuthMiddleware(auther Auther) func(next http.Handler) http.Handler{
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("user_jwt")
			if err != nil {
				http.Error(w, "not found cookie", http.StatusUnauthorized)
				return
			}

			id, err := auther.GetId(r.Context(), cookie.Value)
			if err != nil {
				http.Error(w, "failed to get id from auth", http.StatusUnauthorized)
				return
			}

			ctx := logger.AppendCtx(r.Context(), UserIdKey, id)
			ctx = context.WithValue(ctx, UserIdKey, id)
			r = r.WithContext(ctx)
			
			next.ServeHTTP(w, r)
		})
	}
}
