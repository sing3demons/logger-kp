package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/sing3demons/logger-kp/logger"
)

func main() {
	logger.LoadLogConfig(logger.LogConfig{
		Summary: logger.SummaryLogConfig{
			LogFile:    false,
			LogConsole: false,
		},
		Detail: logger.DetailLogConfig{
			LogFile: false,
		},
	})

	logg := logger.NewLogger()

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		logger.InitSession(r.Context(), logg)

		l := logger.NewLog(r.Context())
		l.Info("test")
		node := "client"
		session := "session:" + uuid.New().String()
		detailLog := logger.NewDetailLog(session, "", "root")
		summaryLog := logger.NewSummaryLog(session, "", "root")
		detailLog.AddInputHttpRequest(node, "get", "invoke", r, false)
		summaryLog.AddSuccess(node, "get", "invoke", "success")

		data := map[string]interface{}{"message": "success"}

		detailLog.AddOutputRequest(node, "get", "invoke", data, data)

		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(data)
		detailLog.End()
		summaryLog.End("", "success")
	})

	http.ListenAndServe(":8080", nil)
}
