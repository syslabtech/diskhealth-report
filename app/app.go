package app

import (
	"bufio"
	"disktoolhealth/config"
	"disktoolhealth/core"
	"disktoolhealth/storage/kafka"
	"disktoolhealth/storage/logging"
	"disktoolhealth/util/metrics"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	batchSize     = 1
	sourceTopic   = config.GetString("Kafka.Topics.SourceTopic")
	metricsTopic  = config.GetString("Kafka.Topics.MetricsTopic")
	consumerGroup = config.GetString("Kafka.ConsumerGroup")
	metricsObject = metrics.PrometheusObject()
)

func StartApp() {

	_ = reversedDataTransformation()
	// fmt.Println(deviceInfo)

	// processBatch()
}

func processBatch() {

	messages, errConsumeMes := consumeKafkaMessage(batchSize)
	if errConsumeMes != nil {
		fmt.Println("Error consuming messages: ", errConsumeMes)
	} else {
		// fmt.Println(messages.Host.Hostname)
		// fmt.Println(messages.Message)

		// Original DiskTool Data
		err := reverseTransformation(messages.Message)
		if err != nil {
			logging.DoLoggingLevelBasedLogs(logging.Error, "", err)
		} else {
			deviceInfo := reversedDataTransformation()
			log.Println(deviceInfo)
			calculationDeviceHealth(deviceInfo)
		}
	}

	fmt.Println("sleep...")
	time.Sleep(10 * time.Second)
	processBatch()

}

// consume the message from the kafka and return the customer id list
func consumeKafkaMessage(batch int) (diskToolMessage core.DiskToolData, kafkaError error) {
	result, errConsume := kafka.Consume(sourceTopic, consumerGroup, batch)
	if errConsume != nil {
		kafkaError = errConsume
	} else {
		for _, msg := range result.Messages {
			json.Unmarshal(msg.Value, &diskToolMessage)
			if strings.HasPrefix("Drive Scanning Output", diskToolMessage.Message) {
				return diskToolMessage, nil
			}
		}
	}
	return
}

func reverseTransformation(modifiedOutput string) (err error) {
	// Replace '||' with newline characters
	temp := "TEMP_NEWLINE_REPLACEMENT"
	modifiedOutput = strings.ReplaceAll(modifiedOutput, "|", temp)

	// Restore '\r\n' using the temporary replacement
	restoredOutput := strings.ReplaceAll(modifiedOutput, temp, "\n")

	// Return the reversed output
	err = os.WriteFile(core.DeviceInfoFile, []byte(restoredOutput), 0644)
	if err != nil {
		return logging.EnrichErrorWithStackTrace(errors.New("Error writing to file: %v\n" + err.Error()))
	}
	return
}

func reversedDataTransformation() (deviceInfo core.DeviceInfo) {
	// Open the file
	file, err := os.Open(core.DeviceInfoFile)
	if err != nil {
		logging.DoLoggingLevelBasedLogs(logging.Error, "", logging.EnrichErrorWithStackTrace(errors.New("Error on opening file: "+err.Error())))
		return
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "On ") && strings.Contains(line, ", Output for") {
			prefix := "On "
			suffix := ", Output for"

			// Find the start and end index of the hostname
			startIndex := strings.Index(line, prefix) + len(prefix)
			endIndex := strings.Index(line[startIndex:], suffix)

			// Adjust endIndex to be relative to the entire line
			if endIndex != -1 {
				endIndex += startIndex // Convert relative index to absolute index
				deviceInfo.Hostname = line[startIndex:endIndex]
			}

			prefix = "Output for "
			suffix = " on Mount"

			// Find the start and end index of the device path
			startIndex = strings.Index(line, prefix) + len(prefix)
			endIndex = strings.Index(line[startIndex:], suffix)

			// Adjust endIndex to be relative to the entire line
			if endIndex != -1 {
				endIndex += startIndex // Convert relative index to absolute index
				deviceInfo.DevicePath = line[startIndex:endIndex]
			}

			prefix = "on Mount:"
			suffix = ":"

			// Find the start and end index of the device path
			startIndex = strings.Index(line, prefix) + len(prefix)
			endIndex = strings.Index(line[startIndex:], suffix)

			// Adjust endIndex to be relative to the entire line
			if endIndex != -1 {
				endIndex += startIndex // Convert relative index to absolute index
				deviceInfo.MountPath = line[startIndex:endIndex]
			}

		} else if strings.Contains(line, "Model Family") {
			deviceInfo.ModelFamily = strings.TrimSpace(strings.Split(line, ":")[1])
		} else if strings.Contains(line, "Device Model") {
			deviceInfo.DeviceModel = strings.TrimSpace(strings.Split(line, ":")[1])
		} else if strings.Contains(line, "Model Number") {
			deviceInfo.ModelNumber = strings.TrimSpace(strings.Split(line, ":")[1])
		} else if strings.Contains(line, "Serial Number") {
			deviceInfo.SerialNumber = strings.TrimSpace(strings.Split(line, ":")[1])
		} else if strings.Contains(line, "Power_On_Hours") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {

				deviceInfo.PowerOnHours, _ = strconv.Atoi(removeCommas(fields[9]))

			}
		} else if strings.Contains(line, "Power On Hours") {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				deviceInfo.PowerOnHours, _ = strconv.Atoi(removeCommas(fields[3]))
			}
		} else if strings.Contains(line, "Host_Writes_32MiB") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				deviceInfo.HostWrites32MiB, _ = strconv.Atoi(removeCommas(fields[9]))
			}
		} else if strings.Contains(line, "Lifetime_Writes_GiB") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				deviceInfo.LifetimeWritesGiB, _ = strconv.Atoi(removeCommas(fields[9]))
			}
		} else if strings.Contains(line, "Total_LBAs_Written") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				deviceInfo.TotalLBAsWritten, _ = strconv.Atoi(removeCommas(fields[9]))
			}
		} else if strings.Contains(line, "Data Units Written") {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				deviceInfo.DataUnitsWritten, _ = strconv.Atoi(removeCommas(fields[3]))
			}
		} else if strings.Contains(line, "Media_Wearout_Indicator") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				deviceInfo.MediaWearoutIndicator, _ = strconv.Atoi(removeCommas(fields[9]))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Output the extracted information
	fmt.Println("Model Family:", deviceInfo.ModelFamily)
	fmt.Println("Device Model:", deviceInfo.DeviceModel)
	fmt.Println("Model Number:", deviceInfo.ModelNumber)
	fmt.Println("Serial Number:", deviceInfo.SerialNumber)
	fmt.Println("Power On Hours:", deviceInfo.PowerOnHours)
	fmt.Println("Host Writes (32MiB):", deviceInfo.HostWrites32MiB)
	fmt.Println("Lifetime Writes (GiB):", deviceInfo.LifetimeWritesGiB)
	fmt.Println("Total LBAs Written:", deviceInfo.TotalLBAsWritten)
	fmt.Println("Data Units Written:", deviceInfo.DataUnitsWritten)
	fmt.Println("Media Wearout Indicator:", deviceInfo.MediaWearoutIndicator)
	fmt.Println("Device Path:", deviceInfo.DevicePath)
	fmt.Println("Mount Path:", deviceInfo.MountPath)
	fmt.Println("Hostname:", deviceInfo.Hostname)

	calculationDeviceHealth(deviceInfo)

	return

}

// Function to remove commas from numeric strings
func removeCommas(value string) string {
	return strings.ReplaceAll(value, ",", "")
}

func calculationDeviceHealth(data core.DeviceInfo) {

	bytesPerBlock := 33554432
	dataUnit := 512

	if data.HostWrites32MiB != 0 {

		//  15200 is TBW for Serial No PHYG2363002W3P8EGN
		totalWrittenData,reamaningLifeDays := calculationData(data.PowerOnHours, data.HostWrites32MiB, bytesPerBlock, deviceTBW(data.SerialNumber))
        
		// produce metrics
		metricsObject["power_on_hour"].WithLabelValues(data.DeviceModel,data.SerialNumber,data.DevicePath,data.MountPath).Set(float64(data.PowerOnHours))
		metricsObject["total_written_data"].WithLabelValues(data.DeviceModel,data.SerialNumber,data.DevicePath,data.MountPath).Set(float64(totalWrittenData))
		metricsObject["remaning_life_in_days"].WithLabelValues(data.DeviceModel,data.SerialNumber,data.DevicePath,data.MountPath).Set(float64(reamaningLifeDays))

		logging.DoLoggingLevelBasedLogs(logging.Debug, "Disk: "+data.DevicePath+", On Host: "+data.Hostname+", Remaning LifeSpan in Days: "+fmt.Sprintf("%.0f", reamaningLifeDays), nil)
	} else if data.DataUnitsWritten != 0 {
		// Data Units Written
		// TBW 21850
		totalWrittenData, reamaningLifeDays := calculationData(data.PowerOnHours, data.DataUnitsWritten, dataUnit, deviceTBW(data.SerialNumber))
		
		// produce metrics
		metricsObject["power_on_hour"].WithLabelValues(data.DeviceModel,data.SerialNumber,data.DevicePath,data.MountPath).Set(float64(data.PowerOnHours))
		metricsObject["total_written_data"].WithLabelValues(data.DeviceModel,data.SerialNumber,data.DevicePath,data.MountPath).Set(float64(totalWrittenData))
		metricsObject["remaning_life_in_days"].WithLabelValues(data.DeviceModel,data.SerialNumber,data.DevicePath,data.MountPath).Set(float64(reamaningLifeDays))

		logging.DoLoggingLevelBasedLogs(logging.Debug, "Disk: "+data.DevicePath+", On Host: "+data.Hostname+", Remaning LifeSpan in Days: "+fmt.Sprintf("%.0f", reamaningLifeDays), nil)

	} else if data.TotalLBAsWritten != 0 {
		totalWrittenData, reamaningLifeDays := calculationData(data.PowerOnHours, data.TotalLBAsWritten, dataUnit, deviceTBW(data.SerialNumber))
		
		// produce metrics
		metricsObject["power_on_hour"].WithLabelValues(data.DeviceModel,data.SerialNumber,data.DevicePath,data.MountPath).Set(float64(data.PowerOnHours))
		metricsObject["total_written_data"].WithLabelValues(data.DeviceModel,data.SerialNumber,data.DevicePath,data.MountPath).Set(float64(totalWrittenData))
		metricsObject["remaning_life_in_days"].WithLabelValues(data.DeviceModel,data.SerialNumber,data.DevicePath,data.MountPath).Set(float64(reamaningLifeDays))

		logging.DoLoggingLevelBasedLogs(logging.Debug, "Disk: "+data.DevicePath+", On Host: "+data.Hostname+", Remaning LifeSpan in Days: "+fmt.Sprintf("%.0f", reamaningLifeDays), nil)

	} else if data.MediaWearoutIndicator != 0 {
		totalWrittenData, reamaningLifeDays := calculationData(data.PowerOnHours, data.MediaWearoutIndicator, dataUnit, deviceTBW(data.SerialNumber))
		
		// produce metrics
		metricsObject["power_on_hour"].WithLabelValues(data.DeviceModel,data.SerialNumber,data.DevicePath,data.MountPath).Set(float64(data.PowerOnHours))
		metricsObject["total_written_data"].WithLabelValues(data.DeviceModel,data.SerialNumber,data.DevicePath,data.MountPath).Set(float64(totalWrittenData))
		metricsObject["remaning_life_in_days"].WithLabelValues(data.DeviceModel,data.SerialNumber,data.DevicePath,data.MountPath).Set(float64(reamaningLifeDays))

		logging.DoLoggingLevelBasedLogs(logging.Debug, "Disk: "+data.DevicePath+", On Host: "+data.Hostname+", Remaning LifeSpan in Days: "+fmt.Sprintf("%.0f", reamaningLifeDays), nil)
	}
}

func calculationData(power, unitWritten, unitSize, tbw int) (totalWrittenData, remaningLifeDaysOfDevice float32) {

	powerOnOneDay := float32(power) / 24
	log.Printf("powerOnOneDay: %.2f \n", powerOnOneDay)

	totalWrittenData = float32(unitWritten*unitSize) / 1024 / 1024 / 1024
	log.Printf("totalWrittenData: %.2f \n", totalWrittenData)

	totalLifeOfDevice := tbw * 1000
	log.Println("totalLifeOfDevice: ", totalLifeOfDevice)

	remaningLifeOfDevice := float32(totalLifeOfDevice) - totalWrittenData
	log.Printf("remaningLifeOfDevice: %.2f \n", remaningLifeOfDevice)

	dailyWrittenData := totalWrittenData / powerOnOneDay
	log.Printf("dailyWrittenData: %.2f \n", dailyWrittenData)

	remaningLifeDaysOfDevice = remaningLifeOfDevice / dailyWrittenData
	log.Printf("remaningLifeDaysOfDevice: %.2f", remaningLifeDaysOfDevice)

	return

}

func deviceTBW(serialNo string) (tbw int) {
	data, err := os.ReadFile(core.DeviceTBWFile)
	if err != nil {
		logging.DoLoggingLevelBasedLogs(logging.Error, "", logging.EnrichErrorWithStackTrace(err))
	}
	// Declare a map to hold the JSON data
	var jsonData map[string]int

	// Unmarshal the JSON data into the map
	if err := json.Unmarshal(data, &jsonData); err != nil {
		logging.DoLoggingLevelBasedLogs(logging.Error, "", logging.EnrichErrorWithStackTrace(err))
	}

	// Lookup the key in the map and return the value
	if value, exists := jsonData[serialNo]; exists {
		return value
	}
	return
}
