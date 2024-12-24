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
			LogFile:    true,
			LogConsole: false,
		},
		Detail: logger.DetailLogConfig{
			LogFile: true,
		},
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
