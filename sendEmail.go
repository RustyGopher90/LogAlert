package main

import (
	"encoding/base64"
	"fmt"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func SendEmailAlert(matches []string, savedFileMatches []string, loglocation LogLocation, fileLocation string, configValues Config) error {
	LogInfo(fmt.Sprintf("Sending emails to : %v ", strings.Join(loglocation.SmtpRecipients, ",")))
	serverAddr := fmt.Sprintf("%v:%v", configValues.SMTPAddress, configValues.SMTPPort)
	from := configValues.SMTPSender
	subject, err := filepath.Abs(fileLocation)
	if err != nil {
		LogInfo(err.Error())
	}
	var body strings.Builder
	for _, match := range matches {
		coloredString := ColorMatch(match, loglocation.SearchTerms)
		body.WriteString(coloredString + "<br>" + "<br>")
	}
	for _, match := range savedFileMatches {
		coloredString := ColorMatch(match, loglocation.SearchTerms)
		body.WriteString(coloredString + "<br>" + "<br>")
	}

	client, err := smtp.Dial(serverAddr)
	if err != nil {
		return err
	}

	defer client.Close()

	if err = client.Mail(from); err != nil {
		return err
	}

	for _, recipient := range loglocation.SmtpRecipients {
		if err = client.Rcpt(recipient); err != nil {
			return err
		}
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}

	message := "To: " + strings.Join(loglocation.SmtpRecipients, ",") + "\r\n" +
		"From: " + from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"Content-Transfer-Encoding: base64\r\n" +
		"\r\n" + base64.StdEncoding.EncodeToString([]byte(string(body.String())))

	_, err = writer.Write([]byte(message))
	if err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return err
	}
	client.Quit()
	return nil
}

func CheckFileLength(fileLocation string) int {
	_, err := os.Stat(fileLocation)
	if os.IsNotExist(err) {
		return 0
	}
	filePath, err := filepath.Abs(fileLocation)
	if err != nil {
		LogInfo(err.Error())
	}
	file, err := os.Stat(filePath)
	if err != nil {
		LogInfo(err.Error())
	}
	return int(file.Size())
}

func RetrySendEmailAlert(matches []string, savedFileMatches []string, loglocation LogLocation, fileLocation string, configValues Config, sentinel int) {
	if sentinel >= 5 {
		LogInfo("Writing matches to file.")
		err := WriteMatchesToFile(filepath.Base(fileLocation), matches)
		if err != nil {
			LogInfo("ERROR: " + err.Error())
		}
		return
	}
	time.Sleep(time.Minute * 2)
	err := SendEmailAlert(matches, savedFileMatches, loglocation, fileLocation, configValues)
	if err != nil {
		LogInfo("ERROR: " + err.Error())
		sentinel++
		RetrySendEmailAlert(matches, savedFileMatches, loglocation, fileLocation, configValues, sentinel)
	}
}
