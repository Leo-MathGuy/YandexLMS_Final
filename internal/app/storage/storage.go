package storage

import (
	"crypto/sha256"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/logging"
	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/processing"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// MARK: DB
var db *sql.DB

func ConnectDB() {
	database, err := sql.Open("sqlite3", "./sqlite3.db")
	if err != nil {
		logging.Panic("DB Failed")
	}
	db = database
}

func query(q string) (sql.Result, error) {
	if db == nil {
		logging.Error("Database not connected")
		return nil, fmt.Errorf("database not connected")
	}
	if res, err := db.Exec(q); err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

func createTables() {
	if _, err := query(`
	CREATE TABLE IF NOT EXISTS Users (
	id INT AUTO_INCREMENT PRIMARY KEY
	login VARCHAR(32) NOT NULL UNIQUE
	passHash BLOB(32) NOT NULL
	)
	`); err != nil {
		logging.Panic(err.Error())
	}
}

// MARK: Users
type User struct {
	login    string
	passHash [32]byte
}

type Users struct {
	users map[string]*User
	sync.RWMutex
}

var U Users

func (u *Users) AddUser(login, password string) {
	u.Lock()
	defer u.Unlock()
	p := &User{login, sha256.Sum256([]byte(password))}
	u.users[strings.ToLower(login)] = p
}

func (u *Users) UserExists(login string) (exists bool) {
	u.RLock()
	defer u.RUnlock()

	_, exists = u.users[strings.ToLower(login)]
	return exists
}

func (u *Users) CheckPass(login, password string) (correct bool) {
	u.RLock()
	defer u.RUnlock()

	return reflect.DeepEqual(sha256.Sum256([]byte(password)), u.users[strings.ToLower(login)].passHash)
}

// MARK: JWT
var secretKey = []byte("ultra-secret-key")

func CreateToken(username string) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"user": strings.ToLower(username),
			"nbf":  now.Add(time.Minute).Unix(),
			"exp":  now.Add(30 * time.Minute).Unix(),
			"iat":  now.Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// MARK: Expressions
type Expression struct {
	rootNode *processing.Node
	done     bool
}

type Expressions struct {
	exprs map[string]*Expression
}

var E Expressions

func (e *Expressions) AddExpression(rootNode *processing.Node) (id string) {
	id = uuid.NewString()
	e.exprs[id] = &Expression{rootNode, false}
	return id
}

// MARK: Tasks
type Task struct {
	pNode *processing.Node
	left  float64
	right float64
	op    rune

	value *float64
	done  bool
}

type Tasks struct {
	tasks []*Task
}

var T Tasks

func (e *Tasks) AddTask(node *processing.Node, left float64, right float64) {
	t := Task{
		node, left, right, *node.Op, nil, false,
	}

	e.tasks = append(e.tasks, &t)
}

func Init() {
	U = Users{}
	U.users = make(map[string]*User)

	E = Expressions{}
	E.exprs = make(map[string]*Expression)

	T = Tasks{}
	T.tasks = make([]*Task, 0)
}
