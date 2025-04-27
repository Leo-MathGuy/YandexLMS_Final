package web

import (
	"net/http"

	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/logging"
	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/web/handlers"
)

func createServer() *http.ServeMux {
	logging.Log("Creating server\n")
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.Index)

	return mux
}

func RunServer() {
	mux := createServer()

	logging.Log("Server starting, press enter to stop\n")
	logging.Panic("Server failed with error: %s\n", http.ListenAndServe(":8080", mux).Error())
}
