package main

import (
    "fmt"
    "bytes"
    "encoding/json"
    "net/http"
    "time"
    "github.com/shirou/gopsutil/v3/cpu"
    "github.com/shirou/gopsutil/v3/mem"
    "github.com/shirou/gopsutil/v3/net"
)

type Metrics struct {
    DeviceIdentifier string  `json:"deviceIdentifier"`
    CpuUsage         float64 `json:"cpuUsage"`
    RamUsage         float64 `json:"ramUsage"`
    GpuUsage         float64 `json:"gpuUsage"`
    NetworkIn        float64 `json:"networkIn"`
    NetworkOut       float64 `json:"networkOut"`
    Temperature      float64 `json:"temperature"`
}

func collectMetrics() Metrics {

    cpuPercent, _ := cpu.Percent(0, false)
    memStat, _ := mem.VirtualMemory()
    netStats, _ := net.IOCounters(false)

    return Metrics{
        DeviceIdentifier: "dev002",
        CpuUsage:         cpuPercent[0],
        RamUsage:         memStat.UsedPercent,
        GpuUsage:         0,
        NetworkIn:        float64(netStats[0].BytesRecv),
        NetworkOut:       float64(netStats[0].BytesSent),
        Temperature:      0,
    }
}

func sendMetrics(metrics Metrics) {

    jsonData, _ := json.Marshal(metrics)

    http.Post(
        "http://localhost:8080/metrics",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
}

func main() {
	fmt.Println("Entered Main")
	for{
	    metrics := collectMetrics()

	    sendMetrics(metrics)
		fmt.Println("Sent Metrics then sleeping for 2 seconds")
	    time.Sleep(2 * time.Second)
	}
}

