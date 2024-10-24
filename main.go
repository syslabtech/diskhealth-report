package main

import (
	"disktoolhealth/app"
	"disktoolhealth/config"
	"disktoolhealth/core"
	"disktoolhealth/server"
	"disktoolhealth/storage/logging"
	"time"
)

var (
	initialDelayInSeconds = config.GetInt64("InitialDelayInSeconds")
)

func main() {
	core.LastActiveTime = time.Now().UTC()

	time.Sleep(time.Duration(initialDelayInSeconds) * time.Second)

	go app.StartApp()

    logging.DoLoggingLevelBasedLogs(logging.Info, "Start Server...", nil)
	server.RegisterServer()

}
