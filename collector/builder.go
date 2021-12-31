package collector

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
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

var swiftInfo = &SwiftInfo{}

func UpdateSwiftInfo(jsonBytes []byte) {
	err := json.Unmarshal(jsonBytes, swiftInfo)
	if err != nil {
		log.Fatal().Err(err)
	}
}

func GetSwiftInfo() *SwiftInfo {
	return swiftInfo
}
