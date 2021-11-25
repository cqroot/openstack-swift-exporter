package collector

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

type SwiftInfo struct {
	Account []struct {
		Host string
		Port string
	}
	Container []struct {
		Host string
		Port string
	}
	Object []struct {
		Host    string
		Port    string
		Devices []string
	}
}

func GetSwiftInfo() *SwiftInfo {
	configPath := "/etc/swift_exporter/cluster.json"
	logrus.Debug("Read SwiftInfo from " + configPath)

	file, err := os.Open(configPath)
	if err != nil {
		logrus.Fatal(err)
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		logrus.Fatal(err)
	}

	var si SwiftInfo
	err = json.Unmarshal([]byte(content), &si)
	if err != nil {
		logrus.Fatal(err)
	}
	return &si
}
