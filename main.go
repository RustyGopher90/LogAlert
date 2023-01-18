package main

import (
    "bufio"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io/fs"
    "io/ioutil"
    "net/smtp"
    "os"
    "path/filepath"
    "regexp"
    "runtime"
    "strconv"
    "strings"
    "time"
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
 
func main() {
    for {
        arg := CheckAndReturnArgs(os.Args)
        if arg == "" {
            os.Exit(0)
        }
 
        LogInfo("Reading Json config file")
        jsonFile, err := ioutil.ReadFile(arg)
        if err != nil {
            LogInfo(err.Error())
            os.Exit(0)
        }
 
        var configValues Config
 
        if err := json.Unmarshal(jsonFile, &configValues); err != nil {
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