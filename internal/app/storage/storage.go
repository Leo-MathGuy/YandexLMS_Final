package storage

import (
	"crypto/sha256"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/logging"
	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/processing"
	"github.com/golang-jwt/jwt/v5"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// MARK: TYPES
type User struct {
	id       uint
	login    string
	passhash []byte
}

type Expression struct {
	ID       uint
	UID      uint
	Expr     string
	Result   float64
	Finished bool

	Gen  *processing.Node
	Done []float64
}

type Expressions struct {
	E map[uint]*Expression
	sync.RWMutex
}

// MARK: DB
var D *sql.DB
var defaultDb string = "sqlite3.db"

// Connect DB to the public D variable
func ConnectDB() {
	var database *sql.DB
	if s := os.Getenv("APPDB"); s != "" {
		database, _ = sql.Open("sqlite3", "./sqlite3.db")
	} else {
		database, _ = sql.Open("sqlite3", defaultDb)
	}

	// Check connection
	err := database.Ping()
	if err != nil {
		logging.Panic("DB Failed womp womp")
	}
	D = database
}

func CreateTables(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS Users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	login VARCHAR(32) NOT NULL UNIQUE,
	passhash BLOB(32) NOT NULL
	);

	CREATE TABLE IF NOT EXISTS Expressions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	uid INT NOT NULL,
	expr STRING NOT NULL,
	result DOUBLE DEFAULT 0,
	done BOOLEAN DEFAULT FALSE
	)
	`)
	return err
}

func DropTables(db *sql.DB) {
	if _, err := db.Exec(`
	DROP TABLE Users;
	DROP TABLE Expressions
	`); err != nil {
		logging.Panic(err.Error())
	}
}

// MARK: USERS
func AddUser(db *sql.DB, login, password string) error {
	passhash := sha256.Sum256([]byte(password))
	username := strings.ToLower(login)
	if _, err := db.Exec("INSERT INTO Users (login, passhash) VALUES (?, ?)", username, passhash[:]); err != nil {
		return err
	}
	return nil
}

func GetUsers(db *sql.DB) ([]*User, error) {
	result := make([]*User, 0)
	rows, err := db.Query("SELECT * FROM Users")
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var curUser User
		err := rows.Scan(&curUser.id, &curUser.login, &curUser.passhash)
		if err != nil {
			break
		}
		result = append(result, &curUser)
	}
	return result, nil
}

func GetUser(db *sql.DB, login string) (*User, error) {
	var u User
	row := db.QueryRow("SELECT * FROM Users WHERE login = ?", strings.ToLower(login))
	err := row.Scan(&u.id, &u.login, &u.passhash)
	if err != nil {
		return &User{}, err
	}

	return &u, nil
}

func UserExists(db *sql.DB, login string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM Users WHERE login = ?", strings.ToLower(login)).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to get row: %s", err.Error())
	}
	return count > 0, nil
}

func CheckPass(db *sql.DB, login, password string) (bool, error) {
	user, err := GetUser(db, login)
	if err != nil {
		return false, err
	}
	return reflect.DeepEqual(sha256.Sum256([]byte(password)), user.passhash), nil
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
var E Expressions = Expressions{make(map[uint]*Expression), sync.RWMutex{}}

func LoadExpression(e *Expressions, id, uid uint, str string, result float64, done bool) error {
	if sep, err := processing.Separate([]rune(str)); err != nil {
		return err
	} else if tokens, err := processing.Tokenize(sep); err != nil {
		return err
	} else if eval, err := processing.Eval(tokens, processing.NodeGen); err != nil {
		return err
	} else {
		e.Lock()
		defer e.Unlock()
		expr := &Expression{id, uid, str, result, done, eval, make([]float64, 0)}
		e.E[id] = expr
	}
	return nil
}

func AddExpression(e *Expressions, db *sql.DB, uid uint, str string) error {
	if _, err := db.Exec("INSERT INTO Expressions (uid, expr) VALUES (?, ?)", uid, str); err != nil {
		return err
	}

	r := db.QueryRow("SELECT (id) FROM Expressions order by rowid desc LIMIT 1", uid, str)

	var id uint
	r.Scan(&id)
	if err := LoadExpression(e, id, uid, str, 0, false); err != nil {
		return err
	}

	return r.Err()
}

// Load expressions from db - generating node trees may take some time
func LoadExpressions(db *sql.DB, e *Expressions) error {
	rows, err := db.Query("SELECT * FROM Expressions")
	if err != nil {
		return err
	}
	for rows.Next() {
		var expr Expression
		if err := rows.Scan(&expr.ID, &expr.UID, &expr.Expr, &expr.Result, &expr.Finished); err != nil {
			return err
		}
		if err := LoadExpression(e, expr.ID, expr.UID, expr.Expr, expr.Result, expr.Finished); err != nil {
			return err
		}
	}
	return nil
}

func GetExpressionResult(e *Expressions, id uint) *float64 {
	e.RLock()
	defer e.RUnlock()
	expr := e.E[id]
	var result float64
	if expr.Finished {
		result = expr.Result
	}
	return &result
}
