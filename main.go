package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	for {
		arg := CheckAndReturnArgs(os.Args)
		if arg == "" {
			os.Exit(0)
		}

		configValues, err := GetConfig(arg)
		if err != nil {
			LogInfo(err.Error())
			os.Exit(0)
		}

		for _, loglocation := range configValues.LogLocations {
			var file string

			if strings.Contains(loglocation.FileLocation, "{{{") {
				file = ParseLocationPlaceholder(loglocation.FileLocation)
				if file == "" {
					LogInfo(fmt.Sprintf("Improper date format in config file for %v.", loglocation.FileLocation))
					os.Exit(0)
				}
			} else {
				file = loglocation.FileLocation
			}

			if CheckSearchAndIgnoreDuplicates(loglocation.SearchTerms, loglocation.IgnoreTerms, file) {
				os.Exit(0)
			}

			FileLine := ReadPlaceHolderFile(file)
			currentFileSize := CheckFileLength(file)
			if currentFileSize < FileLine {
				FileLine = 0
			}
			matchesArr, endingline := ReadFileForMatches(file, loglocation.SearchTerms, loglocation.IgnoreTerms, FileLine)
			if len(matchesArr) == 0 && endingline == 0 {
				continue
			}
			savedFileMatches, err := GetSavedFileMatches(filepath.Base(file))
			if err != nil {
				LogInfo("ERROR: " + err.Error())
			}
			if len(matchesArr) > 0 || len(savedFileMatches) > 0 {
				LogInfo(fmt.Sprintf("Found %v matches", len(matchesArr)))
				err := SendEmailAlert(matchesArr, savedFileMatches, loglocation, file, configValues)
				if err != nil {
					RetrySendEmailAlert(matchesArr, savedFileMatches, loglocation, file, configValues, 0)
					LogInfo(err.Error())
				} else {
					err := ClearMatchesFile(filepath.Base(file))
					if err != nil {
						LogInfo("ERROR: " + err.Error())
					}
				}
			} else {
				LogInfo("Found 0 matches")
			}
			WritePlaceHolderFile(file, endingline)
		}

		CleanUpPlaceHolderFiles()
		MinutesToSleep, err := strconv.Atoi(configValues.TimeToSleep)
		if err != nil {
			LogInfo(err.Error())
			os.Exit(0)
		}
		time.Sleep(time.Duration(MinutesToSleep) * time.Minute)
	}
}

func LogInfo(s string) {
	dateTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("%v : %v \n", dateTime, s)
}
