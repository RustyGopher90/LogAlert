package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

func ReadFileForMatches(fileLocation string, searchTerms []string, ignoreList []string, startingpoint int) ([]string, int) {
	LogInfo(fmt.Sprintf("Reading %v for matches.", fileLocation))
	var matchArr []string
	filePath, err := filepath.Abs(fileLocation)
	if err != nil {
		LogInfo(err.Error())
	}
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		LogInfo(err.Error())
		return matchArr, 0
	}

	openSesame, err := os.Open(filePath)
	if err != nil {
		LogInfo(err.Error())
	}
	defer openSesame.Close()

	if _, err := openSesame.Seek(int64(startingpoint), 0); err != nil {
		LogInfo(err.Error())
	}
	scanner := bufio.NewScanner(openSesame)

	endingPoint := startingpoint
	scanLines := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		advance, token, err = bufio.ScanLines(data, atEOF)
		endingPoint += int(advance)
		return
	}
	scanner.Split(scanLines)

	for scanner.Scan() {
		if FindMatch(scanner.Text(), searchTerms) {
			matchArr = append(matchArr, scanner.Text())
		}
	}
	finalArr := FilterMatches(matchArr, ignoreList)
	return finalArr, endingPoint
}

func FindMatch(line string, terms []string) bool {
	for _, s := range terms {
		regexpression, err := regexp.Compile(strings.ToLower(s))
		if err != nil {
			LogInfo(err.Error())
		}
		if regexpression.MatchString(strings.ToLower(line)) {
			return true
		}
	}
	return false
}

func FilterMatches(matchesArr []string, ignoreList []string) []string {
	var finalArr []string
	for _, match := range matchesArr {
		if !FindMatch(match, ignoreList) {
			finalArr = append(finalArr, match)
		}
	}
	return finalArr
}

func ColorMatch(match string, searchTerms []string) string {
	var coloredArr []string
	arr := strings.Split(match, " ")
	for _, a := range arr {
		if FindMatch(a, searchTerms) {
			coloredArr = append(coloredArr, fmt.Sprintf("<span style="+"color:#EC5C46"+" border>%v</span>", a))
		} else {
			coloredArr = append(coloredArr, a)
		}
	}
	return strings.Join(coloredArr, " ")
}

func MakeMatchesFolder() (string, error) {
	if runtime.GOOS == "windows" {
		windowsFolderLocation := "./matches"
		_, err := os.Stat(windowsFolderLocation)
		if os.IsNotExist(err) {
			err := os.Mkdir(windowsFolderLocation, 0664)
			if err != nil {
				return "", err
			}
			return windowsFolderLocation, nil
		}
		return windowsFolderLocation, nil
	} else {
		linuxFolderLocation := "/usr/local/bin/inhouse/logalert/matches"
		_, err := os.Stat(linuxFolderLocation)
		if os.IsNotExist(err) {
			err := os.Mkdir(linuxFolderLocation, 0776)
			if err != nil {
				return linuxFolderLocation, err
			}
			return linuxFolderLocation, nil
		}
		return linuxFolderLocation, nil
	}
}

func ClearMatchesFile(fileName string) error {
	var path string
	if runtime.GOOS == "windows" {
		path = "./matches/"
	} else {
		path = "/usr/local/bin/inhouse/logalert/matches/"
	}
	filePath := fmt.Sprintf("%v%v", path, fileName)
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil
	}
	err = os.Remove(filePath)
	if err != nil {
		return err
	}
	return nil
}

func GetSavedFileMatches(fileName string) ([]string, error) {
	var path string
	var savedFileMatches []string
	if runtime.GOOS == "windows" {
		path = "./matches/"
	} else {
		path = "/usr/local/bin/inhouse/logalert/matches/"
	}
	filePath := fmt.Sprintf("%v%v", path, fileName)
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return []string{}, nil
	} else {
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return []string{}, err
		}
		matches := strings.Split(string(fileData), "\r\n")
		for _, match := range matches {
			if match == "" {
				continue
			}
			savedFileMatches = append(savedFileMatches, match)
		}
		return savedFileMatches, nil
	}

}

func WriteMatchesToFile(fileName string, matches []string) error {
	folderLocation, err := MakeMatchesFolder()
	fileLocation := fmt.Sprintf("%v/%v", folderLocation, fileName)
	if err != nil {
		return err
	}
	openedFile, err := os.OpenFile(fileLocation, os.O_APPEND|os.O_CREATE, 0776)
	if err != nil {
		return err
	}
	defer openedFile.Close()
	for _, match := range matches {
		_, err = openedFile.Write([]byte(match + "\r\n"))
		if err != nil {
			return err
		}
	}
	return nil
}

func CheckSearchAndIgnoreDuplicates(sterm []string, iterm []string, fileLocation string) bool {
	stermMap := make(map[string]string)
	for _, s := range sterm {
		stermMap[s] = s
	}
	for _, ignoreTerm := range iterm {
		if _, ok := stermMap[ignoreTerm]; ok {
			LogInfo("There is an ignore term that is also in the search terms for file location: " + fileLocation + ". Check the config.json file.")
			return true
		}
	}
	return false
}
