package main

import (
	"bufio"
	"context"
	"os"
	"sync"
	"time"

	"github.com/Leo-MathGuy/YandexLMS_Final/cmd/app/web"
	"github.com/Leo-MathGuy/YandexLMS_Final/internal/agent"
	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/logging"
	pc "github.com/Leo-MathGuy/YandexLMS_Final/internal/app/web/grpc"
)

func waitForEnter(wg *sync.WaitGroup) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	wg.Done()
}

func main() {
	if _, err := os.Stat("web"); os.IsNotExist(err) {
		panic("run main in main directory please")
	}

	if logger, err := logging.CreateLogger(); err != nil {
		panic("logger failed to initialize: " + err.Error())
	} else {
		logging.Logger = logger
	}

	go web.RunServer()
	go pc.StartServer(":5050")

	time.Sleep(time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	agent.StartThreads(ctx)

	end := sync.WaitGroup{}
	end.Add(1)
	go waitForEnter(&end)
	end.Wait()
	logging.Log("Exiting...")
	cancel()
	time.Sleep(2 * time.Second)
	logging.Log("Quit.\n")
}
