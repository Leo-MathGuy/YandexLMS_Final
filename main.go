package main

import (
	"bufio"
	"os"
	"sync"

	"github.com/Leo-MathGuy/YandexLMS_Final/cmd/app/web"
	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/logging"
)

func waitForEnter(wg *sync.WaitGroup) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	wg.Done()
}

func main() {
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		panic("run main in main directory please")
	}

	if logger, err := logging.CreateLogger(); err != nil {
		panic("logger failed to initialize: " + err.Error())
	} else {
		logging.Logger = logger
	}

	go web.RunServer()

	end := sync.WaitGroup{}
	end.Add(1)
	go waitForEnter(&end)
	end.Wait()
	logging.Log("Quit.\n")
}
