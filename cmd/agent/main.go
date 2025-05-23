package main

import (
	"bufio"
	"context"
	"os"

	"github.com/Leo-MathGuy/YandexLMS_Final/internal/agent"
	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/logging"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	conn := agent.StartThreads(ctx)
	defer conn.Close()

	go func() {
		logging.Log("Agent started. Press enter to stop")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		cancel()
	}()

	for {
		select {}
	}
}
