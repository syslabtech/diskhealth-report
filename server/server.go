package server

import (
	"disktoolhealth/config"
	"disktoolhealth/healthcheck"
	"disktoolhealth/storage/logging"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var port = config.GetInt("HttpServer.Port")

func RegisterServer() {

	s := chi.NewRouter()

	s.Get("/statichealthcheck", staticHealthCheck())
	s.Get("/healthcheck", healthcheck.HealthCheck())
	s.Handle("/metrics", promhttp.Handler())
	httpPort := ":" + strconv.Itoa(port)
	err_conn := http.ListenAndServe(httpPort, s)
	if err_conn != nil {
		logging.DoLoggingLevelBasedLogs(logging.Error, "", logging.EnrichErrorWithStackTrace(errors.New("http connection error:"+err_conn.Error())))
	}
}

func staticHealthCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logging.DoLoggingLevelBasedLogs(logging.Debug, "healthcheck success", nil)
		w.Write([]byte("Helthcheck Success"))
	}
}
