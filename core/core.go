package core

import "time"

var LastActiveTime time.Time

var DeviceInfoFile string = "outputdiskdata.txt"

var DeviceTBWFile string = "deviceTBW.json"

type DiskToolData struct {
	Host struct {
		Hostname string `json:"hostname"`
	} `json:"host"`
	Message string `json:"message"`
}

type DeviceInfo struct {
	ModelFamily           string
	DeviceModel           string
	ModelNumber           string
	SerialNumber          string
	PowerOnHours          int
	HostWrites32MiB       int
	LifetimeWritesGiB     int
	TotalLBAsWritten      int
	DataUnitsWritten      int
	MediaWearoutIndicator int
	DevicePath            string
	MountPath             string
	Hostname              string
}
