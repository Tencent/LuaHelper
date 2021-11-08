package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

var (
	debugLog *log.Logger
	errorLog *log.Logger
	LspLog   *log.Logger
)

type logger = func(string, ...interface{})

var GFileLog *os.File = nil

// logFlag 为true表示开启日志
func InitLog(logFlag bool) {
	if !logFlag {
		debugLog = nil
		errorLog = nil
		LspLog = nil
		log.SetFlags(0)
		return
	}

	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)
	fileLog, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)

	if err != nil {
		log.Fatalln("failed to open log file:", err)
		return
	}

	debugLog = log.New(io.MultiWriter(fileLog), "Debug ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	errorLog = log.New(io.MultiWriter(fileLog), "Error ", log.Ldate|log.Lmicroseconds|log.Lshortfile)

	LspLog = debugLog
	GFileLog = fileLog
}

// 是否关闭日志
func CloseLog() {
	if GFileLog != nil {
		GFileLog.Close()
	}
}

func Debug(format string, v ...interface{}) {
	if debugLog == nil {
		return
	}

	debugLog.Output(2, fmt.Sprintf(format, v...))
}

func Error(format string, v ...interface{}) {
	if errorLog == nil {
		return
	}

	errorLog.Output(2, fmt.Sprintf(format, v...))
}
