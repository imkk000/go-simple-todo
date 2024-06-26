package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.TimeOnly})
	if len(os.Args) <= 1 {
		err := errors.New("empty task")
		handleErr(err, "invalid arguments")
	}
	home, err := os.UserHomeDir()
	handleErr(err, "get user home")
	path := filepath.Join(home, ".todo.json")
	handleErr(read(path), "read tasks")
	action := strings.ToLower(os.Args[1])
	switch action {
	case "c":
		Create()
	case "l":
		GetAll()
	case "g":
		Get()
	case "u":
		Update()
	case "d":
		Delete()
	default:
		err := fmt.Errorf("unknown action %s", action)
		handleErr(err, "invalid action")
	}
	handleErr(write(path), "write tasks")
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

var tasks []string

func Create() {
	handleValidLength(2)
	task := strings.Join(os.Args[2:], " ")
	tasks = append(tasks, task)

	log.Info().
		Int("index", len(tasks)-1).
		Str("task", task).
		Msg("task created")
}

func GetAll() {
	handleValidLength(1)
	for i, task := range tasks {
		log.Info().
			Int("index", i).
			Str("task", task).
			Msg("list tasks")
	}
}

func Get() {
	handleValidLength(2)
	i, err := strconv.Atoi(os.Args[2])
	handleErr(err, "invalid index")
	handleOutOfLength(i)

	log.Info().
		Int("index", i).
		Str("task", tasks[i]).
		Msg("get task")
}

func Update() {
	handleValidLength(2)
	i, err := strconv.Atoi(os.Args[2])
	handleErr(err, "invalid index")
	handleOutOfLength(i)

	task := strings.Join(os.Args[3:], " ")
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
	tasks = append(tasks[:i], tasks[i+1:]...)

	log.Info().
		Int("index", i).
		Str("task", task).
		Msg("delete task")
}
