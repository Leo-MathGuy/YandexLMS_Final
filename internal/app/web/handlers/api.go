package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/logging"
	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/storage"
)

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data AuthRequest
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Cannot read JSON", http.StatusBadRequest)
		return
	}

	if x, _ := regexp.MatchString("^[A-Za-z][A-Za-z0-9_-]{2,31}$", data.Login); !x {
		http.Error(w, "Invalid username", http.StatusBadRequest)
		return
	}

	if len(data.Password) < 3 || len(data.Password) > 32 {
		http.Error(w, "Invalid password", http.StatusBadRequest)
		return
	}

	if storage.U.UserExists(data.Login) {
		http.Error(w, "User exists", http.StatusBadRequest)
		return
	}

	logging.Log("User created: %s\n", data.Login)
	storage.U.AddUser(data.Login, data.Password)
	w.WriteHeader(200)
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data AuthRequest
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Cannot read JSON", http.StatusBadRequest)
		return
	}

	if !storage.U.UserExists(data.Login) {
		http.Error(w, "User does not exist", http.StatusBadRequest)
		return
	}

	if !storage.U.CheckPass(data.Login, data.Password) {
		http.Error(w, "Wrong passowrd", http.StatusUnauthorized)
		return
	}

	jwt, err := storage.CreateToken(data.Login)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		logging.Error("Failure to create JWT token: %s\n", err.Error())
		return
	}
	w.Write([]byte(jwt))
}
