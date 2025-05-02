package web

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/logging"
	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/storage"
	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/web/handlers"
)

func createServer() *mux.Router {
	logging.Log(strings.Repeat("-", 80) + "\n")
	logging.Log("Creating server\n")

	mux := mux.NewRouter()

	mux.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))

	mux.HandleFunc("/api/v1/register", handlers.RegisterAPI).Methods("POST")
	mux.HandleFunc("/api/v1/login", handlers.LoginAPI).Methods("POST")

	mux.HandleFunc("/api/v1/calculate", handlers.Calculate).Methods("POST")

	mux.HandleFunc("/calc", handlers.Calc).Methods("GET")
	mux.HandleFunc("/login", handlers.Login).Methods("GET")
	mux.HandleFunc("/register", handlers.Register).Methods("GET")
	mux.HandleFunc("/", handlers.Index)
	return mux
}

func initServer() {
	storage.ConnectDB()
	storage.CreateTables(storage.D)
}

func RunServer() {
	mux := createServer()
	initServer()

	logging.Log("Server starting, press enter to stop\n")
	logging.Panic("Server failed with error: %s\n", http.ListenAndServe(":8080", mux).Error())
}
