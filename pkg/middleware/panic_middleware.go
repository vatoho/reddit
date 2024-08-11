package middleware

import (
	"go.uber.org/zap"
	"net/http"

	"reddit/pkg/response"
)

func Panic(logger *zap.SugaredLogger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Infof("panic middleware start")
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf("panic recovered: %s", err)
				response.WriteResponse(logger, w, []byte(`{"message": "panic occurred"}`), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
