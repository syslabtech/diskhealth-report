package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

func PrometheusObject() map[string]*prometheus.GaugeVec {
	
	obj := make(map[string]*prometheus.GaugeVec)

	powerOnHour := getMetricsObject("disktool_power_on_hour")
	obj["power_on_hour"] = powerOnHour

	writtenData := getMetricsObject("disktool_total_written_data")
	obj["total_written_data"] = writtenData

	remaningLifeInDays := getMetricsObject("disktool_remaning_life_in_days")
    obj["remaning_life_in_days"] = remaningLifeInDays
	
	unregisterMetrics()
	return obj
}

func getMetricsObject(metricsName string) *prometheus.GaugeVec {

	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: metricsName ,  // "disktool_health_gauge"
			Help: "DiskTool Health metrics",
		},
		[]string{"device_model_no", "device_serial_no", "device_path", "mount_path"},
	)

	prometheus.MustRegister(gauge)

	return gauge

}

func unregisterMetrics() {
	prometheus.Unregister(collectors.NewGoCollector())
	prometheus.Unregister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
}
