package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

const (
	path = ".todo.yaml"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	home, err := os.UserHomeDir()
	handleErr(err, "get user home")
	path := filepath.Join(home, path)
	handleErr(read(path), "read tasks")

	rootCmd := &cli.Command{
		Version:               "0.0.1",
		EnableShellCompletion: true,
	}
	cmds := []struct {
		Name   string
		Alias  string
		Fn     func(ctx context.Context, c *cli.Command) error
		Before func(ctx context.Context, c *cli.Command) (context.Context, error)
	}{
		{"create", "c", Create, validArgs},
		{"list", "l", GetAll, nil},
		{"get", "g", Get, validArgs},
		{"update", "u", Update, validArgs},
		{"delete", "d", Delete, validArgs},
	}
	for _, cmd := range cmds {
		rootCmd.Commands = append(rootCmd.Commands, &cli.Command{
			Name:    cmd.Name,
			Aliases: []string{cmd.Alias},
			Before:  cmd.Before,
			Action: func(ctx context.Context, c *cli.Command) error {
				handleErr(cmd.Fn(ctx, c), "execute command")
				handleErr(write(path), "write tasks")

				return nil
			},
		})
	}
	if err := rootCmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal().Err(err).Msg("run command")
	}
}

func write(path string) error {
	content, err := yaml.Marshal(tasks)
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
	return yaml.NewDecoder(fs).Decode(&tasks)
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

func validArgs(ctx context.Context, c *cli.Command) (context.Context, error) {
	if c.Args().Len() == 0 {
		return nil, cli.Exit("invalid parameters", 1)
	}
	return ctx, nil
}

func Create(ctx context.Context, c *cli.Command) error {
	defer GetAll(ctx, c)

	task := strings.Join(c.Args().Slice(), " ")
	handleEmptyInput(task)
	tasks = append(tasks, task)

	log.Info().
		Int("index", len(tasks)-1).
		Str("task", task).
		Msg("task created")

	return nil
}

func GetAll(ctx context.Context, c *cli.Command) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.Configure(func(config *tablewriter.Config) {
		config.Row.Formatting.Alignment = tw.AlignLeft
	})
	defer table.Close()

	table.Header("ID", "Task")
	for i, task := range tasks {
		table.Append([]any{i, task})
	}
	table.Render()

	return nil
}

func Get(ctx context.Context, c *cli.Command) error {
	i, err := strconv.Atoi(c.Args().First())
	handleErr(err, "invalid index")
	handleOutOfLength(i)

	fmt.Println(i, tasks[i])

	return nil
}

func Update(ctx context.Context, c *cli.Command) error {
	defer GetAll(ctx, c)

	i, err := strconv.Atoi(c.Args().First())
	handleErr(err, "invalid index")
	handleOutOfLength(i)

	task := strings.Join(c.Args().Tail(), " ")
	handleEmptyInput(task)
	prev := tasks[i]
	tasks[i] = strings.ReplaceAll(task, "@@", prev)

	log.Info().
		Int("index", i).
		Str("from", prev).
		Str("to", tasks[i]).
		Msg("update task")

	return nil
}

func Delete(ctx context.Context, c *cli.Command) error {
	defer GetAll(ctx, c)

	i, err := strconv.Atoi(c.Args().First())
	handleErr(err, "invalid index")
	handleOutOfLength(i)

	task := tasks[i]
	tasks = slices.Delete(tasks, i, i+1)

	log.Info().
		Int("index", i).
		Str("task", task).
		Msg("delete task")

	return nil
}
