package status

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// UptimeRecord menyimpan data waktu saat node pertama kali dinyalakan
type UptimeRecord struct {
	StartTime int64 `json:"start_time"`
}

var UptimeFile = "eterbit_data/uptime.json"

// RecordStartTime dipanggil saat node booting untuk mencatat waktu mulai
func RecordStartTime() {
	os.MkdirAll("eterbit_data", 0755)
	
	// Jika file uptime sudah ada (artinya node mungkin sudah jalan), jangan Timpa kecuali node baru restart
	if _, err := os.Stat(UptimeFile); os.IsNotExist(err) {
		record := UptimeRecord{
			StartTime: time.Now().Unix(),
		}
		data, _ := json.MarshalIndent(record, "", "  ")
		os.WriteFile(UptimeFile, data, 0644)
	}
}

// GetUptime menghitung berapa lama node telah aktif (dalam detik, menit, jam, hari) mirip dogecoin-cli uptime
func GetUptime() (int64, string) {
	data, err := os.ReadFile(UptimeFile)
	if err != nil {
		// Jika file belum ada (node belum pernah jalan/booting lewat daemon), catat waktu sekarang
		StartTime := time.Now().Unix()
		record := UptimeRecord{StartTime: StartTime}
		newData, _ := json.MarshalIndent(record, "", "  ")
		os.MkdirAll("eterbit_data", 0755)
		os.WriteFile(UptimeFile, newData, 0644)
		return 0, "0 seconds"
	}

	var record UptimeRecord
	json.Unmarshal(data, &record)

	now := time.Now().Unix()
	diff := now - record.StartTime
	if diff < 0 {
		diff = 0
	}

	// Konversi detik ke format human-readable (Hari, Jam, Menit, Detik)
	days := diff / 86400
	hours := (diff % 86400) / 3600
	minutes := (diff % 3600) / 60
	seconds := diff % 60

	var result string
	if days > 0 {
		result = fmt.Sprintf("%d days, %d hours, %d minutes, %d seconds", days, hours, minutes, seconds)
	} else if hours > 0 {
		result = fmt.Sprintf("%d hours, %d minutes, %d seconds", hours, minutes, seconds)
	} else if minutes > 0 {
		result = fmt.Sprintf("%d minutes, %d seconds", minutes, seconds)
	} else {
		result = fmt.Sprintf("%d seconds", seconds)
	}

	return diff, result
}
