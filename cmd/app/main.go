package main

import (
	"bufio"
	"os"
	"sync"

	"github.com/Leo-MathGuy/YandexLMS_Final/cmd/app/web"
	pc "github.com/Leo-MathGuy/YandexLMS_Final/internal/app/web/grpc"
)

func waitForEnter(wg *sync.WaitGroup) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	wg.Done()
}

func main() {
	go web.RunServer()
	go pc.StartServer(":5050")
	wg := sync.WaitGroup{}
	wg.Add(1)
	go waitForEnter(&wg)
	wg.Wait()
}
