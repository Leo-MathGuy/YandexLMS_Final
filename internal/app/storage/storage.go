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
	ID       uint
	Login    string
	Passhash []byte
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
func ConnectDB() chan struct{} {
	var database *sql.DB
	if s := os.Getenv("APPDB"); s != "" {
		database, _ = sql.Open("sqlite3", s)
	} else {
		database, _ = sql.Open("sqlite3", defaultDb)
	}

	// Check connection
	err := database.Ping()
	if err != nil {
		logging.Panic("DB Failed womp womp")
	}
	D = database

	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				err := database.Ping()
				if err != nil {
					logging.Panic("DB Failed this gets 5 big booms, boom boom boom boom boom")
				}
			}
		}
	}()

	return stop
}

func DisconnectDB() {
	D.Close()
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
		err := rows.Scan(&curUser.ID, &curUser.Login, &curUser.Passhash)
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
	err := row.Scan(&u.ID, &u.Login, &u.Passhash)
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
	hash := sha256.Sum256([]byte(password))
	return reflect.DeepEqual(hash[:], user.Passhash), nil
}

// MARK: JWT
var secretKey = []byte("ultra-secret-key")

func CreateToken(username string) (string, error) {
	now := time.Now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"user": strings.ToLower(username),
			"exp":  now.Add(30 * time.Minute).UnixMilli(),
			"iat":  now.UnixMilli(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func CheckToken(db *sql.DB, token string) (*User, error) {
	t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return secretKey, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		return nil, err
	}

	if claims, ok := t.Claims.(jwt.MapClaims); ok {
		now := time.Now().UTC()
		expire := time.UnixMilli(int64(claims["exp"].(float64)))
		if now.After(expire) {
			return nil, nil
		} else if user, err := GetUser(db, claims["user"].(string)); err != nil {
			return nil, err
		} else {
			return user, nil
		}
	} else {
		return nil, fmt.Errorf("token invalid")
	}
}

// MARK: Expressions
var E Expressions = Expressions{make(map[uint]*Expression), sync.RWMutex{}}

func LoadExpression(e *Expressions, id, uid uint, str string, result float64, done bool) error {
	var eval *processing.Node

	if done {
		eval = nil
	} else if sep, err := processing.Separate([]rune(str)); err != nil {
		return err
	} else if tokens, err := processing.Tokenize(sep); err != nil {
		return err
	} else if eval, err = processing.Eval(tokens, processing.NodeGen); err != nil {
		return err
	}

	e.Lock()
	defer e.Unlock()
	expr := &Expression{id, uid, str, result, done, eval, make([]float64, 0)}
	e.E[id] = expr

	return nil
}

func AddExpression(e *Expressions, db *sql.DB, uid uint, str string) (uint, error) {
	if _, err := db.Exec("INSERT INTO Expressions (uid, expr) VALUES (?, ?)", uid, str); err != nil {
		return 0, err
	}

	r := db.QueryRow("SELECT (id) FROM Expressions order by rowid desc LIMIT 1", uid, str)

	var id uint = 0
	r.Scan(&id)
	if err := LoadExpression(e, id, uid, str, 0, false); err != nil {
		db.Exec("DELETE FROM Expressions WHERE id=?", id) // Attempt to clear - no error checking here
		return 0, err
	}

	if err := r.Err(); err != nil {
		return 0, err
	} else if id == 0 {
		return 0, fmt.Errorf("what")
	} else {
		return id, nil
	}
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
