package middleware

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"strings"

	"reddit/pkg/response"

	"reddit/pkg/session"
)

type userKey int

const MyUserKey userKey = 1

func Auth(logger *zap.SugaredLogger, sm *session.SessionManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Infof("auth middleware start")
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			response.WriteResponse(logger, w, []byte(`{"message": "there is no access token or it is in wrong format"}`),
				http.StatusUnauthorized)
			return
		}
		tokenValue := strings.TrimPrefix(authHeader, "Bearer ")
		mySession, err := sm.GetSession(tokenValue)
		if err != nil || mySession == nil {
			errText := fmt.Sprintf(`{"message": "there is no session for token %s}`, tokenValue)
			response.WriteResponse(logger, w, []byte(errText), http.StatusUnauthorized)
			return
		}
		sessionUser := mySession.User
		ctx := r.Context()
		ctx = context.WithValue(ctx, MyUserKey, sessionUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
