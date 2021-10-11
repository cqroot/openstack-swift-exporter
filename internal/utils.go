package internal

import (
	"encoding/json"
	"net"
	"net/http"
	"time"
)

// Check port connectivity. If connected, return 1, otherwise return 0.
func checkPort(host string, port string) float64 {
	timeout := 10 * time.Second
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		return 0
	}
	if conn != nil {
		defer conn.Close()
	}
	return 1
}

func getDiskUsage(host string, port string) ([]map[string]interface{}, error) {
	url := "http://" + host + ":" + port + "/recon/diskusage"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&result)
	return result, err
}
