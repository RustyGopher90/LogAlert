package main

import (
	"encoding/json"
	"io/ioutil"
	"strings"
)

type Config struct {
	TimeToSleep  string        `json:"minutesToSleep"`
	SMTPAddress  string        `json:"smtpAddress"`
	SMTPPort     string        `json:"smtpPort"`
	SMTPSender   string        `json:"smtpSender"`
	LogLocations []LogLocation `json:"logLocations"`
}

type LogLocation struct {
	FileLocation   string   `json:"fileLocation"`
	SmtpRecipients []string `json:"smtpRecipients"`
	SearchTerms    []string `json:"searchTerms"`
	IgnoreTerms    []string `json:"ignoreTerms"`
}

func CheckAndReturnArgs(args []string) string {
	if len(args) < 2 {
		LogInfo("You need to pass a json config file! \nExample : ./logalertgo.exe ./config.json")
		return ""
	}
	if len(args) != 2 {
		LogInfo("only one argument can be passed to the application. The argument needs to be a json config file! \nExample : ./logalertgo.exe ./config.json")
		return ""
	}
	if !strings.Contains(args[1], ".json") {
		LogInfo("You can only pass a json file")
		return ""
	}
	return args[1]
}
func GetConfig(filepath string) (Config, error) {
	LogInfo("Reading Json config file")
	var configValues Config
	jsonFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		return configValues, err
	}

	if err := json.Unmarshal(jsonFile, &configValues); err != nil {
		return configValues, err
	}
	return configValues, nil
}
