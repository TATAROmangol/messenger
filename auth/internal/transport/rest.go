package transport

import (
	"time"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"auth_service/pkg/postgres"
	"auth_service/pkg/logger"
	"encoding/json"
	"go.uber.org/zap"
	"context"
	"crypto/md5"
	"encoding/hex"
)

func Register(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie(JWTCookieName)
	if err == http.ErrNoCookie {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn(r.Context(), "Error in JSON decoder", zap.Error(err))
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if req.Login == "" || req.Pass == "" {
			logger.Warn(r.Context(), "Error: required credential[s] are empty", zap.Error(err))
			http.Error(w, "Empty credentials", http.StatusBadRequest)
			return
		}

		claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": req.Login,
			"exp": time.Now().Add(7*24*time.Hour).Unix(), // Неделя - "7 раз по 24 часа"
			"iat": time.Now().Unix(),
		})
		stringKey, err := claims.SignedString([]byte(JWTKey))
		if err != nil {
			logger.Warn(r.Context(), "Failed to create JWT token", zap.Error(err))
			http.Error(w, "Failed to create JWT token", http.StatusInternalServerError)
			return
		}

		passHash := md5.Sum([]byte(req.Pass))
		pass := hex.EncodeToString(passHash[:])
		_, err = postgres.InsertUser(stringKey, req.Login, req.Email, pass, req.Name) // TODO: куда отдавать айди?
		if err != nil {
			logger.Warn(r.Context(), "Failed to insert to the database", zap.Error(err))
			http.Error(w, "Failed to insert to the database", http.StatusInternalServerError)
			return
		}

		// http.SetCookie(w, &http.Cookie{
		// 	Name:    JWTCookieName,
		// 	Value:   stringKey,
		// 	Expires: time.Now().Add(7*24*time.Hour),
		// 	Path:    "/",
		// 	Domain:   "", 
		// 	HttpOnly: true,
		// 	SameSite: http.SameSiteLaxMode,
		// })
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(RegisterResponse{Token: stringKey}); err != nil {
			logger.Warn(r.Context(), "Error in JSON encoder", zap.Error(err))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
	// TODO: если уже зареган, redirect...
}

func Login(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie(JWTCookieName)
	if err == http.ErrNoCookie {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn(r.Context(), "Error in JSON decoder", zap.Error(err))
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}


		passHash := md5.Sum([]byte(req.Pass))
		pass := hex.EncodeToString(passHash[:])
		stringKey, err := postgres.GetTokenByCredentialAndPassword(req.Credential, pass)
		if err != nil {
			logger.Warn(r.Context(), "Failed to get user record", zap.Error(err))
			http.Error(w, "Failed to get user record", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    JWTCookieName,
			Value:   stringKey,
			Expires: time.Now().Add(7*24*time.Hour),
			Path:    "/",
			Domain:   "", 
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(LoginResponse{Token: stringKey}); err != nil {
			logger.Warn(r.Context(), "Error in JSON encoder", zap.Error(err))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie(JWTCookieName)
	if err == http.ErrNoCookie {
		logger.Info(r.Context(), "ErNoCookie when logging out", zap.Error(err))
		http.Error(w, "Can not logout without logging in!", http.StatusBadRequest)
		return
	}
	if err != nil {
		logger.Info(r.Context(), "Failed to get cookie", zap.Error(err))
		http.Error(w, "Failed to get cookie", http.StatusInternalServerError) // TODO: InternalError ли?
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    JWTCookieName,
		Value:   "",
		Expires: time.Now().Add(-1*time.Hour),
		Path:    "/",
		Domain:   "", 
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	// TODO: redirect...
}

func Delete(w http.ResponseWriter, r *http.Request) {
	tokenCookie, err := r.Cookie(JWTCookieName)
	if err == http.ErrNoCookie {
		logger.Info(r.Context(), "ErrNoCookie when deleting user", zap.Error(err))
		http.Error(w, "Can not delete user without logging in!", http.StatusBadRequest)
		return
	}
	if err != nil {
		logger.Info(r.Context(), "Failed to get cookie", zap.Error(err))
		http.Error(w, "Failed to get cookie", http.StatusInternalServerError) // TODO: InternalError ли?
		return
	}
	err = postgres.DeleteUserByToken((*tokenCookie).Value)
	if err != nil {
		logger.Warn(r.Context(), "Failed to delete user", zap.Error(err))
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}
	Logout(w, r)
}

func MiddlewareHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		guid := uuid.New().String()
		ctx, _ := logger.New(r.Context())
		ctx = context.WithValue(ctx, logger.RequestIDKey, guid)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}