package web

import (
	"assignment4/clinic"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"
)

func doLog(req *http.Request, logType, msg string) {

	file, err := os.OpenFile("./logs/out.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}
	defer file.Close()

	var logger *log.Logger

	logType = strings.ToUpper(logType)

	if logType == "INFO" {
		logger = log.New(io.MultiWriter(os.Stdout, file), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	} else if logType == "WARNING" {
		logger = log.New(io.MultiWriter(os.Stderr, file), "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	} else if logType == "ERROR" {
		logger = log.New(io.MultiWriter(os.Stderr, file), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		logger = log.New(ioutil.Discard, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	}

	logger.Println(req.RemoteAddr, msg)
}

func time2HumanReadable(t int64) string {
	return time.Unix(t, 0).Format("3:04PM")
}

func getUserByID(uid string) *clinic.Patient {
	user, err := clinic.GetPatientByID(uid)

	if err == nil {
		return user
	}

	return nil
}

func ucFirst(str string) string {
	if len(str) == 0 {
		return ""
	}
	tmp := []rune(str)
	tmp[0] = unicode.ToUpper(tmp[0])
	return string(tmp)
}

func stripSpace(str string) string {
	return strings.ReplaceAll(str, " ", "")
}
