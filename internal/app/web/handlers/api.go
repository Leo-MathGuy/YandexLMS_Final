package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/logging"
	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/processing"
	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/storage"
	"github.com/gorilla/mux"
)

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func RegisterAPI(w http.ResponseWriter, r *http.Request) {
	var data AuthRequest
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Cannot read JSON", http.StatusBadRequest)
		return
	}

	if x, _ := regexp.MatchString("^[A-Za-z][A-Za-z0-9_\\-]{2,31}$", data.Login); !x {
		http.Error(w, "Invalid username", http.StatusBadRequest)
		return
	}

	if len(data.Password) < 3 || len(data.Password) > 32 {
		http.Error(w, "Invalid password", http.StatusBadRequest)
		return
	}

	if n, err := storage.UserExists(storage.D, data.Login); err != nil {
		logging.Error("Failed getting user exists: %s", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	} else if n {
		http.Error(w, "User exists", http.StatusBadRequest)
		return
	}

	logging.Log("User created: %s\n", data.Login)
	storage.AddUser(storage.D, data.Login, data.Password)
	w.WriteHeader(200)
}

func LoginAPI(w http.ResponseWriter, r *http.Request) {
	var data AuthRequest
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Cannot read JSON", http.StatusBadRequest)
		return
	}

	if n, err := storage.UserExists(storage.D, data.Login); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		logging.Error("Failed login: %s", err.Error())
	} else if !n {
		http.Error(w, "User does not exist", http.StatusBadRequest)
		return
	}

	if n, err := storage.CheckPass(storage.D, data.Login, data.Password); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		logging.Error("Failed login: %s", err.Error())
	} else if !n {
		http.Error(w, "Wrong password", http.StatusUnauthorized)
		return
	}

	jwt, err := storage.CreateToken(data.Login)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		logging.Error("Failure to create JWT token: %s\n", err.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    jwt,
		Path:     "/",
		MaxAge:   1800,
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	w.Write([]byte(jwt))
}

type CalcRequest struct {
	Expr  string `json:"expression"`
	Token string `json:"token"`
}

type CalcResponse struct {
	ID uint `json:"id"`
}

func Calculate(w http.ResponseWriter, r *http.Request) {
	var data CalcRequest
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Cannot read JSON", http.StatusBadRequest)
		return
	}

	var u *storage.User

	if user, err := storage.CheckToken(storage.D, data.Token); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	} else if user == nil {
		http.Error(w, "Token expired", http.StatusForbidden)
		return
	} else {
		u = user
	}

	validation := processing.Validate(data.Expr)
	if validation != nil {
		http.Error(w, fmt.Sprintf("Invalid expression: %s", validation.Error()), http.StatusBadRequest)
		return
	}

	if id, err := storage.AddExpression(&storage.E, storage.D, u.ID, data.Expr); err != nil {
		logging.Error("Error adding expression: %s", err.Error())
		http.Error(w, "Internal server error", http.StatusBadRequest)
	} else if response, err := json.Marshal(&CalcResponse{id}); err != nil {
		logging.Error("Error marhshaling response: %s", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	} else if err := storage.GenTasks(&storage.T, storage.E.E[id]); err != nil {
		logging.Error("Error generating tasks: %s", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	} else {
		w.Write(response)
	}
}

type ExpressionsResponse struct {
	Exprs []storage.Expression `json:"expressions"`
}

func ExpressionsAPI(w http.ResponseWriter, r *http.Request) {
	authToken := r.Header.Get("Authentication")
	if authToken == "" {
		http.Error(w, "Please log in", http.StatusForbidden)
		return
	}

	user, err := storage.CheckToken(storage.D, authToken)

	if err != nil {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		logging.Error("Error reading token: %s", err.Error())
		return
	}

	if user == nil {
		http.Error(w, "Invalid token", http.StatusForbidden)
		return
	}

	result, err := storage.GetExpressionsForUser(storage.D, user)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		logging.Error("Error finding expressions: %s", err.Error())
		return
	}
	unpointered := make([]storage.Expression, 0)
	for _, v := range result {
		unpointered = append(unpointered, *v)
	}

	response, err := json.Marshal(ExpressionsResponse{unpointered})
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		logging.Error("Error marhshaling response: %s", err.Error())
		return
	}

	w.Write(response)
}

type ExpressionResponse struct {
	Expr storage.Expression `json:"expression"`
}

func ExpressionAPI(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, _ := strconv.Atoi(idStr)

	authToken := r.Header.Get("Authentication")
	if authToken == "" {
		http.Error(w, "Please log in", http.StatusForbidden)
		return
	}

	user, err := storage.CheckToken(storage.D, authToken)

	if err != nil {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		logging.Error("Error reading token: %s", err.Error())
		return
	}

	if user == nil {
		http.Error(w, "Invalid token", http.StatusForbidden)
		return
	}

	result, err := storage.GetExpressionForUser(storage.D, user, id)
	if err != nil {
		http.Error(w, "No such expression", http.StatusNotFound)
		logging.Log("No expression: %s", err.Error())
		return
	}

	response, err := json.Marshal(ExpressionResponse{result})
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		logging.Error("Error marhshaling response: %s", err.Error())
		return
	}

	w.Write(response)
}

func Favicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/static/favicon.ico")
}
