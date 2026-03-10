package middleware

import (
	"AssetManagement/db/dbHelper"
	"AssetManagement/utils"
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type UserCtx struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	Role      string `json:"role"`
}

type ContextKeyType struct{}

var UserContextKey = ContextKeyType{}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			utils.RespondError(w, nil, "missing token", http.StatusUnauthorized)
			return
		}

		token, parseErr := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method ")
			}
			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})

		if parseErr != nil || !token.Valid {
			utils.RespondError(w, parseErr, "invalid token", http.StatusUnauthorized)
			return
		}

		claimValues, ok := token.Claims.(jwt.MapClaims)

		if !ok {
			utils.RespondError(w, nil, "invalid token claims", http.StatusUnauthorized)
			return
		}
		sessionID := claimValues["session_id"].(string)

		archivedAt, err := dbHelper.GetArchivedAt(sessionID)

		if err != nil {
			utils.RespondError(w, err, "internal server error", http.StatusInternalServerError)
			return
		}

		if archivedAt != nil {
			utils.RespondError(w, nil, "invalid token", http.StatusUnauthorized)
			return
		}

		user := &UserCtx{
			UserID:    claimValues["user_id"].(string),
			SessionID: sessionID,
			Role:      claimValues["role"].(string),
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserContext(r *http.Request) *UserCtx {
	user, _ := r.Context().Value(UserContextKey).(*UserCtx)
	return user
}

func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			userCtx := UserContext(r)
			if userCtx == nil {
				utils.RespondError(w, nil, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims := r.Context().Value(UserContextKey).(*UserCtx)
			role := claims.Role

			for _, allowed := range allowedRoles {
				if role == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}

			utils.RespondError(w, nil, "access denied", http.StatusForbidden)
		})
	}
}
