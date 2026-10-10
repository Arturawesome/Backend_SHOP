package web

import (
	pbAuth "BE2_API_Shop/gen/auth"
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userIDKey contextKey = "userID"

func (app *Application) authJwtMiddleware(next http.Handler) http.Handler {
	handF := func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(
				w,
				http.StatusText(http.StatusUnauthorized),
				http.StatusUnauthorized,
			)
			return
		}
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			http.Error(
				w,
				http.StatusText(http.StatusUnauthorized),
				http.StatusUnauthorized,
			)
			return
		}

		grpcValidateTokenRequest := pbAuth.ValidateTokenRequest{
			Token: tokenString,
		}
		resp, err := app.grpcAuthClient.ValidateToken(r.Context(), &grpcValidateTokenRequest)
		if err != nil || !resp.IsValid {
			http.Error(
				w,
				http.StatusText(http.StatusUnauthorized),
				http.StatusUnauthorized,
			)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, resp.UserId)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handF)
}
