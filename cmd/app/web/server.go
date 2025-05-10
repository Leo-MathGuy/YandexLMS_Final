package web

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"

	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/logging"
	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/storage"
	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/util"
	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/web/handlers"
)

var StopDb chan<- struct{}

func createServer() *mux.Router {
	util.Leave()
	logging.Log(strings.Repeat("-", 80) + "\n")
	logging.Log("Creating server\n")

	mux := mux.NewRouter()

	mux.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))

	mux.HandleFunc("/api/v1/register", handlers.RegisterAPI).Methods("POST")
	mux.HandleFunc("/api/v1/login", handlers.LoginAPI).Methods("POST")

	mux.HandleFunc("/api/v1/calculate", handlers.Calculate).Methods("POST")

	mux.HandleFunc("/favicon.ico", handlers.Favicon)
	mux.HandleFunc("/calc", handlers.Calc).Methods("GET")
	mux.HandleFunc("/login", handlers.Login).Methods("GET")
	mux.HandleFunc("/register", handlers.Register).Methods("GET")
	mux.HandleFunc("/", handlers.Index)
	return mux
}

func initServer() {
	logging.Log("Connecting database")
	stopDb := storage.ConnectDB()
	if err := storage.CreateTables(storage.D); err != nil {
		logging.Panic("Could not create tables: %s", err.Error())
	}

	logging.Log("Checking templates")
	handlers.CheckTemplates()

	logging.Log("Loading expressions")
	if err := storage.LoadExpressions(storage.D, &storage.E); err != nil {
		logging.Panic("Error loading expressions: %s", err.Error())
	}

	logging.Log("Generating tasks")
	if err := storage.GenAllTasks(&storage.T, &storage.E); err != nil {
		logging.Panic("Error generating tasks")
	}
	StopDb = stopDb
}

func RunServer() error {
	connected := true
	if logging.Logger == log.Default() {
		connected = false
		logging.CreateLogger()
		logging.Warning("App launched without agent")
	}

	mux := createServer()
	initServer()

	logging.Log("Server starting on :8080, press enter to stop\n")
	err := http.ListenAndServe(":8080", mux)
	logging.Panic("Server failed with error: %s\n", err.Error())

	if connected {
		fmt.Fprintln(os.Stdin)
	}
	return err
}
