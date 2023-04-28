package logger

import (
	"io"
	"log"
	"os"
)

var (
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
	FatalLogger   *log.Logger
	WarningLogger *log.Logger
	TraceLogger	  *log.Logger
	ClientTraceLogger	  *log.Logger
)

const LogsPath = "./logs/"

func init() {
	if _, err := os.Stat(LogsPath); os.IsNotExist(err) {
		err = os.Mkdir(LogsPath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}
	logFile, err := os.OpenFile(LogsPath+"log.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	tracefile, err := os.OpenFile(LogsPath+"trace.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	InfoLogger = log.New(io.MultiWriter(os.Stdout), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	ErrorLogger = log.New(multiWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	FatalLogger = log.New(multiWriter, "FATAL: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(multiWriter, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	TraceLogger = log.New(tracefile, "GEDS-TRACE; ", 0)
	ClientTraceLogger = log.New(multiWriter, "GEDS-CLIENT-TRACE; ", 0)
}
