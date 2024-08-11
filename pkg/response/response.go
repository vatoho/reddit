package response

import (
	"go.uber.org/zap"
	"net/http"
)

func WriteResponse(logger *zap.SugaredLogger, w http.ResponseWriter, dataJSON []byte, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(statusCode)
	_, err := w.Write(dataJSON)
	if err != nil {
		logger.Errorf("error in writing response body: %s", err)
	}
}
