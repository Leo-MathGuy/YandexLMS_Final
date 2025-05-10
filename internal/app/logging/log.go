package logging

import (
	"io"
	"log"
	"os"
)

var Logger *log.Logger = log.Default()

func CreateLogger() (logger *log.Logger, err error) {
	f, err := os.OpenFile("logs/appLog.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	w := io.MultiWriter(os.Stdout, f)
	logger = log.New(w, "", log.Ldate|log.Ltime|log.Lmicroseconds)

	return logger, nil
}

func Log(s string, f ...any) {
	Logger.Printf(s, f...)
}

func Warning(s string, f ...any) {
	Logger.Printf("WARNING: "+s, f...)
}
func Error(s string, f ...any) {
	Logger.Printf("ERROR: "+s, f...)
}

func Panic(s string, f ...any) {
	Logger.Panicf("PANIC: "+s, f...)
}
