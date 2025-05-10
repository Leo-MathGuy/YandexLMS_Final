package storage

import (
	"crypto/sha256"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strconv"
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
	TaskID   uint
	Finished bool

	Gen *processing.Node
}

type Expressions struct {
	E map[uint]*Expression
	sync.RWMutex
}

// MARK: DB
var D *sql.DB
var defaultDb string = "sqlite3.db"

// Connect DB to the D variable
func ConnectDB() chan<- struct{} {
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
	done BOOLEAN DEFAULT FALSE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
	expr := &Expression{id, uid, str, result, 0, done, eval}
	e.E[id] = expr

	if eval != nil && eval.IsValue {
		expr.Finished = true
		expr.Result = *eval.Value
	}

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
		DeleteExpression(db, e, id) // Attempt to clear - no error checking here
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

func DeleteExpression(db *sql.DB, e *Expressions, id uint) error {
	res, err := db.Exec("DELETE FROM Expressions WHERE id=?", id)

	if r, _ := res.RowsAffected(); r > 1 {
		panic("Database was cleared " + strconv.FormatInt(r, 10))
	}

	return err
}

// Load expressions from db - generating node trees may take some time
func LoadExpressions(db *sql.DB, e *Expressions) error {
	rows, err := db.Query("SELECT id, uid, expr, result, done FROM Expressions")
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

func CheckExpressions(db *sql.DB, e *Expressions, t *Tasks) {
	t.Lock()
	defer t.Unlock()

	e.Lock()
	defer e.Unlock()

	for _, v := range e.E {
		if v.TaskID != 0 && t.T[v.TaskID].Value {
			v.Finished = true
			v.Result = *t.T[v.TaskID].Left
		}
	}
}

// MARK: Tasks
type Tasks struct {
	T    map[uint]*Task
	Keys []uint // To keep earlier tasks higher priority
	Id   uint
	sync.RWMutex
}

type Task struct {
	ID     uint
	Left   *float64
	Right  *float64
	Op     *rune
	Value  bool
	ExprId uint

	LeftT  *Task
	RightT *Task

	Parent *uint
	Node   *processing.Node
	Sent   bool
}

var T Tasks = Tasks{make(map[uint]*Task), make([]uint, 0), 0, sync.RWMutex{}}

func GetTasks(tasks *Tasks, n *processing.Node, parent uint) (*Task, error) {
	var t Task

	tasks.Id++
	myId := tasks.Id

	if n.IsValue {
		t = Task{tasks.Id, n.Value, nil, nil, true, 0, nil, nil, &parent, n, false}
		tasks.Lock()
		defer tasks.Unlock()
		tasks.T[tasks.Id] = &t
		tasks.Keys = append(tasks.Keys, tasks.Id)
		return &t, nil
	}

	t = Task{myId, nil, nil, n.Op, false, 0, nil, nil, &parent, n, false}

	func() {
		tasks.Lock()
		defer tasks.Unlock()
		tasks.T[myId] = &t
		tasks.Keys = append(tasks.Keys, myId)
	}()

	if l, err := GetTasks(tasks, n.Left, myId); err != nil {
		return nil, err
	} else {
		t.LeftT = l
	}

	if r, err := GetTasks(tasks, n.Right, myId); err != nil {
		return nil, err
	} else {
		t.RightT = r
	}

	return &t, nil
}

// Generates bottom level tasks for an expression
func GenTasks(tasks *Tasks, expr *Expression) error {
	if t, err := GetTasks(tasks, expr.Gen, 0); err != nil {
		DeleteTaskRec(tasks, t)
		return err
	} else {
		expr.TaskID = t.ID
		t.ExprId = expr.ID
		FlattenTask(tasks, t)
		return nil
	}
}

func DeleteTask(tasks *Tasks, id uint) {
	delete(tasks.T, id)
	tasks.Keys = slices.DeleteFunc(tasks.Keys, func(x uint) bool { return id == x })
}

func DeleteTaskRec(tasks *Tasks, task *Task) {
	if task.LeftT != nil {
		DeleteTaskRec(tasks, task.LeftT)
	}
	if task.RightT != nil {
		DeleteTaskRec(tasks, task.RightT)
	}

	tasks.Lock()
	defer tasks.Unlock()
	DeleteTask(tasks, task.ID)
}

func FlattenTask(tasks *Tasks, task *Task) {
	if task.LeftT != nil && task.LeftT.Value {
		func() {
			tasks.Lock()
			defer tasks.Unlock()

			task.Left = task.LeftT.Left
			DeleteTask(tasks, task.LeftT.ID)
			task.LeftT = nil
		}()

	} else {
		if task.LeftT != nil {
			FlattenTask(tasks, task.LeftT)
		}
	}

	if task.RightT != nil && task.RightT.Value {
		func() {
			tasks.Lock()
			defer tasks.Unlock()

			task.Right = task.RightT.Left
			DeleteTask(tasks, task.RightT.ID)
			task.RightT = nil
		}()
	} else {
		if task.RightT != nil {
			FlattenTask(tasks, task.RightT)
		}
	}
}

func FinishTask(db *sql.DB, tasks *Tasks, e *Expressions, id uint, result float64) error {
	tasks.Lock()
	defer tasks.Unlock()

	if task, ok := tasks.T[id]; !ok {
		return fmt.Errorf("no such task")
	} else if !task.Sent {
		logging.Error("Not sent task recived: %d", task.ID)
		return fmt.Errorf("task not sent")
	} else if task.Value {
		return fmt.Errorf("task is value/recived already")
	} else if task.LeftT != nil || task.RightT != nil {
		logging.Error("task recieved before expected: %d", task.ID)
		return fmt.Errorf("there are still dependencies")
	} else {
		task.Left = &result
		if *task.Parent != 0 {
			parent := tasks.T[*task.Parent]
			if parent.LeftT == task {
				parent.Left = task.Left
				parent.LeftT = nil
			} else if parent.RightT == task {
				parent.Right = task.Left
				parent.RightT = nil
			} else {
				logging.Error("something wrong with parent: %d", task.ID)
				return fmt.Errorf("something wrong with parent")
			}

			DeleteTask(tasks, id)
		} else {
			e.Lock()
			defer e.Unlock()
			ex := e.E[task.ExprId]
			ex.Finished = true
			ex.Result = result
			ex.TaskID = 0
			_, err := db.Exec("UPDATE Expressions SET done = ?, result = ? WHERE id = ?", true, result, task.ExprId)
			if err != nil {
				return err
			}
			DeleteTask(tasks, task.ID)
		}

		return nil
	}
}

// Gets a task for the agent. nil = no tasks. Sets task as sent
func GetReadyTask(tasks *Tasks) *Task {
	tasks.RLock()
	defer tasks.RUnlock()

	for _, k := range tasks.Keys {
		task := tasks.T[k]

		if task == nil {
			DeleteTask(tasks, k)
		} else if !(task.LeftT != nil || task.RightT != nil || task.Sent) {
			task.Sent = true
			return task
		}
	}

	return nil
}

func GenAllTasks(tasks *Tasks, exprs *Expressions) error {
	exprs.Lock()
	defer exprs.Unlock()
	for _, v := range exprs.E {
		if v.Finished || v.TaskID != 0 {
			continue
		}

		err := GenTasks(tasks, v)
		if err != nil {
			return err
		}
	}
	return nil
}
