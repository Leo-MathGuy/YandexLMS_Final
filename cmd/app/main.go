package main

import (
	"github.com/Leo-MathGuy/YandexLMS_Final/cmd/app/web"
	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/web/grpc"
)

func main() {
	web.RunServer()
	grpc.StartServer(":5050")
}
