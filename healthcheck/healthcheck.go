package healthcheck

import (
	"disktoolhealth/config"
	"disktoolhealth/core"
	"disktoolhealth/storage/logging"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

var maxInActiveTime = config.GetInt("MaxInActiveTime_Minute")

func HealthCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lastActiveTime := core.LastActiveTime
		currentTime := time.Now().UTC()

		diffrence := currentTime.Sub(lastActiveTime).Minutes()

		if int(diffrence) >= maxInActiveTime {

			diffMinute := fmt.Sprintf("%f", diffrence)
			logging.DoLoggingLevelBasedLogs(logging.Error, "", logging.EnrichErrorWithStackTrace(errors.New("Consumer not active from last "+diffMinute+" minutes")))
			giveHealthCheckResponseBasedOnHealth(false, w)
		} else {
			logging.DoLoggingLevelBasedLogs(logging.Info, "healthcheck success", nil)
			giveHealthCheckResponseBasedOnHealth(true, w)
		}
	}
}

func giveHealthCheckResponseBasedOnHealth(healthCheckStatus bool, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	if healthCheckStatus {
		resp["message"] = "Success"
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		resp["message"] = "Failed"
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Println("error json marshal in healthcheck:", err)
	}
	w.Write(jsonResp)
}
