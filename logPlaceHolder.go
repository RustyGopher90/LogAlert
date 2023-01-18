package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func WritePlaceHolderFile(fileLocation string, endingLine int) {
	MakePlaceHolderDir()
	fileLocation = filepath.Base(fileLocation)
	var file *os.File
	var err error
	if runtime.GOOS == "windows" {
		file, err = os.Create(fmt.Sprintf("./FilePlaceHolders/%v", fileLocation))
		if err != nil {
			LogInfo(err.Error())
		}
	} else {
		file, err = os.Create(fmt.Sprintf("/usr/local/bin/inhouse/logalert/FilePlaceHolders/%v", fileLocation))
		if err != nil {
			LogInfo(err.Error())
		}
	}
	_, err = file.WriteString(fmt.Sprintf("%v", endingLine))
	if err != nil {
		LogInfo(err.Error())
	}
	file.Close()
}

func MakePlaceHolderDir() {
	if runtime.GOOS == "windows" {
		_, err := os.Stat("./FilePlaceHolders")
		if os.IsNotExist(err) {
			err := os.Mkdir("./FilePlaceHolders", 0664)
			if err != nil {
				LogInfo(err.Error())
			}
		}
	} else {
		_, err := os.Stat("/usr/local/bin/inhouse/logalert/FilePlaceHolders")
		if os.IsNotExist(err) {
			err := os.Mkdir("/usr/local/bin/inhouse/logalert/FilePlaceHolders", 0776)
			if err != nil {
				LogInfo(err.Error())
			}
		}
	}
}

func ReadPlaceHolderFile(fileLocation string) int {
	var number []byte
	var err error
	if runtime.GOOS == "windows" {
		_, err := os.Stat(fmt.Sprintf("./FilePlaceHolders/%v", filepath.Base(fileLocation)))
		if os.IsNotExist(err) {
			return 0
		}
		number, err = os.ReadFile(fmt.Sprintf("./FilePlaceHolders/%v", filepath.Base(fileLocation)))
		if err != nil {
			LogInfo(err.Error())
		}
	} else {
		_, err := os.Stat(fmt.Sprintf("/usr/local/bin/inhouse/logalert/FilePlaceHolders/%v", filepath.Base(fileLocation)))
		if os.IsNotExist(err) {
			return 0
		}
		number, err = os.ReadFile(fmt.Sprintf("/usr/local/bin/inhouse/logalert/FilePlaceHolders/%v", filepath.Base(fileLocation)))
		if err != nil {
			LogInfo(err.Error())
		}
	}
	fileLineNum, err := strconv.Atoi(string(number))
	if err != nil {
		LogInfo(err.Error())
	}
	return fileLineNum
}

func CleanUpPlaceHolderFiles() {
	var files []fs.FileInfo
	var err error
	var filePlaceHolderLocation string
	if runtime.GOOS == "windows" {
		filePlaceHolderLocation = "./FilePlaceHolders/"
		files, err = ioutil.ReadDir("./FilePlaceHolders")
		if err != nil {
			LogInfo(err.Error())
		}
	} else {
		filePlaceHolderLocation = "/usr/local/bin/inhouse/logalert/FilePlaceHolders/"
		files, err = ioutil.ReadDir("/usr/local/bin/inhouse/logalert/FilePlaceHolders")
		if err != nil {
			LogInfo(err.Error())
		}
	}
	for _, file := range files {
		deleteFile := file.ModTime().Before(time.Now().Add((-7 * 24) * time.Hour))
		if deleteFile {
			err := os.Remove(filePlaceHolderLocation + file.Name())
			if err != nil {
				LogInfo(err.Error())
			}
			LogInfo(fmt.Sprintf("Removed file %v from FilePlaceHolders. It was older than a week, last modTime was: %v", file, file.ModTime()))
		}
	}
}

func ParseLocationPlaceholder(fileLocation string) string {
	replacePlaceHolder := strings.ReplaceAll(fileLocation, "{{{", " ")
	replacePlaceHolder2 := strings.ReplaceAll(replacePlaceHolder, "}}}", " ")
	fileLocationArr := strings.Split(replacePlaceHolder2, " ")
	if fileLocationArr[1] == "yyyyMMdd" {
		currentDate := time.Now().Format("20060102")
		return fmt.Sprintf("%v%v%v", fileLocationArr[0], currentDate, fileLocationArr[2])
	}
	return ""
}
