package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

const baseURL = "http://100.75.228.102:8080"
const apiKeyFile = "apikey.txt"

type Metrics struct {
	DeviceIdentifier string  `json:"deviceIdentifier"`
	CpuUsage         float64 `json:"cpuUsage"`
	RamUsage         float64 `json:"ramUsage"`
	GpuUsage         float64 `json:"gpuUsage"`
	NetworkIn        float64 `json:"networkIn"`
	NetworkOut       float64 `json:"networkOut"`
	Temperature      float64 `json:"temperature"`
}

type Device struct {
	DeviceIdentifier string `json:"deviceIdentifier"`
	DeviceName       string `json:"deviceName"`
	DeviceType       string `json:"deviceType"`
	ApiKey           string `json:"apiKey"`
}

var apiKey string

// 🔹 Read API key from file
func loadApiKey() bool {
	data, err := os.ReadFile(apiKeyFile)
	if err != nil {
		return false
	}

	apiKey = string(data)
	fmt.Println("Loaded API Key:", apiKey)
	return true
}

// 🔹 Save API key to file
func saveApiKey(key string) {
	os.WriteFile(apiKeyFile, []byte(key), 0644)
}

// 🔹 Register device if no key exists
func registerDevice() {

	device := Device{
		DeviceIdentifier: "dev003Rpi",
		DeviceName:       "Raspberry Pi",
		DeviceType:       "Pi",
	}

	jsonData, _ := json.Marshal(device)

	resp, err := http.Post(
		baseURL+"/devices/register",
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		fmt.Println("Registration failed:", err)
		return
	}

	defer resp.Body.Close()

	var registered Device
	json.NewDecoder(resp.Body).Decode(&registered)

	apiKey = registered.ApiKey

	saveApiKey(apiKey)

	fmt.Println("Registered. Saved API Key:", apiKey)
}

// 🔹 Collect system metrics
func collectMetrics() Metrics {

	cpuPercent, _ := cpu.Percent(0, false)
	memStat, _ := mem.VirtualMemory()
	netStats, _ := net.IOCounters(false)

	return Metrics{
		DeviceIdentifier: "dev003Rpi",
		CpuUsage:         cpuPercent[0],
		RamUsage:         memStat.UsedPercent,
		GpuUsage:         0,
		NetworkIn:        float64(netStats[0].BytesRecv),
		NetworkOut:       float64(netStats[0].BytesSent),
		Temperature:      0,
	}
}

// 🔹 Send metrics with API key
func sendMetrics(metrics Metrics) {

	jsonData, _ := json.Marshal(metrics)

	req, _ := http.NewRequest(
		"POST",
		baseURL+"/metrics",
		bytes.NewBuffer(jsonData),
	)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", apiKey)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Send failed:", err)
		return
	}

	fmt.Println("Status:", resp.Status)
	resp.Body.Close()
}

func main() {

	fmt.Println("Starting agent...")

	// 🔹 Try loading existing key
	if !loadApiKey() {
		fmt.Println("No API key found. Registering device...")
		registerDevice()
	}

	// 🔹 Safety check
	if apiKey == "" {
		fmt.Println("No API key. Exiting.")
		os.Exit(1)
	}

	for {

		metrics := collectMetrics()

		sendMetrics(metrics)

		time.Sleep(2 * time.Second)
	}
}
