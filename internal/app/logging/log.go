package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var Logger *log.Logger = log.Default()

const logFilePath = "logs/appLog.log"

func rotateFile() {
	info, err := os.Stat(logFilePath)
	if err != nil || info.Size() <= 1024*64 { // 64KB
		return
	}

	timestamp := time.Now().UTC().Format("20060102T150405")

	baseName := fmt.Sprintf("appLog%s%s", timestamp, ".bak")
	newPath := filepath.Join(filepath.Dir(logFilePath), baseName)

	os.Rename(logFilePath, newPath)
}

func CreateLogger() (logger *log.Logger, err error) {
	rotateFile() // I kinda dont care if this doesnt work

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
