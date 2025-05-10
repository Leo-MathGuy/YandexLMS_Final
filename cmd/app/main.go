package main

import (
	"github.com/Leo-MathGuy/YandexLMS_Final/cmd/app/web"
	pc "github.com/Leo-MathGuy/YandexLMS_Final/internal/app/web/grpc"
)

func main() {
	web.RunServer()
	pc.StartServer(":5050")
}
