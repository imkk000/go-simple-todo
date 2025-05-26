package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	path = ".todo.json"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.TimeOnly})
	if len(os.Args) <= 1 {
		PrintHelp()
		return
	}
	home, err := os.UserHomeDir()
	handleErr(err, "get user home")
	path := filepath.Join(home, path)
	handleErr(read(path), "read tasks")
	action := strings.ToLower(os.Args[1])

	switch action {
	case "h", "help":
		PrintHelp()
	case "c", "create":
		Create()
	case "l", "list":
		GetAll()
	case "g", "get":
		Get()
	case "u", "update":
		Update()
	case "d", "delete":
		Delete()
	default:
		err := fmt.Errorf("unknown action %s", action)
		handleErr(err, "invalid action")
	}
	handleErr(write(path), "write tasks")
}

func PrintHelp() {
	fmt.Println("Usage: todo [action] [args]")
	fmt.Println("Actions:")
	fmt.Println("  c, create   - Create a new task")
	fmt.Println("  l, list     - List all tasks")
	fmt.Println("  g, get      - Get a specific task by index")
	fmt.Println("  u, update   - Update a specific task by index")
	fmt.Println("  d, delete   - Delete a specific task by index")
	fmt.Println("  h, help     - Show this help message")
}

func write(path string) error {
	content, err := json.Marshal(tasks)
	if err != nil {
		return err
	}
	return os.WriteFile(path, content, 0644)
}

func read(path string) (err error) {
	fs, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer func() {
		err = errors.Join(err, fs.Close())
	}()
	return json.NewDecoder(fs).Decode(&tasks)
}

func handleErr(err error, msg string) {
	if err != nil {
		log.Fatal().Err(err).Msg(msg)
	}
}

func handleOutOfLength(i int) {
	if i < 0 || i >= len(tasks) {
		err := errors.New("out of index")
		handleErr(err, "invalid index")
	}
}

func handleValidLength(min int) {
	if l := len(os.Args[1:]); l < min {
		err := fmt.Errorf("must be at least %d arguments", min)
		handleErr(err, "invalid arguments")
	}
}

func handleEmptyInput(input string) {
	if len(input) == 0 {
		err := errors.New("empty task")
		handleErr(err, "invalid input")
	}
}

var tasks []string

func Create() {
	handleValidLength(2)
	task := strings.Join(os.Args[2:], " ")
	handleEmptyInput(task)
	tasks = append(tasks, task)

	log.Info().
		Int("index", len(tasks)-1).
		Str("task", task).
		Msg("task created")
}

func GetAll() {
	handleValidLength(1)
	for i, task := range tasks {
		fmt.Println(i, task)
	}
}

func Get() {
	handleValidLength(2)
	i, err := strconv.Atoi(os.Args[2])
	handleErr(err, "invalid index")
	handleOutOfLength(i)

	fmt.Println(i, tasks[i])
}

func Update() {
	handleValidLength(2)
	i, err := strconv.Atoi(os.Args[2])
	handleErr(err, "invalid index")
	handleOutOfLength(i)

	task := strings.Join(os.Args[3:], " ")
	handleEmptyInput(task)

	prev := tasks[i]
	tasks[i] = task

	log.Info().
		Int("index", i).
		Str("from", prev).
		Str("to", tasks[i]).
		Msg("update task")
}

func Delete() {
	handleValidLength(2)
	i, err := strconv.Atoi(os.Args[2])
	handleErr(err, "invalid index")
	handleOutOfLength(i)

	task := tasks[i]
	tasks = slices.Delete(tasks, i, i+1)

	log.Info().
		Int("index", i).
		Str("task", task).
		Msg("delete task")
}
