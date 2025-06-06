package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

const (
	path = ".todo.yaml"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:             os.Stdout,
		FormatTimestamp: func(any) string { return "" },
	})

	home, err := os.UserHomeDir()
	handleErr(err, "get user home")
	path := filepath.Join(home, path)
	handleErr(read(path), "read tasks")

	rootCmd := &cli.Command{
		Version:               "0.0.2",
		EnableShellCompletion: true,
		DefaultCommand:        "list",
		Commands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"l"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Value:   "table",
					},
				},
				Action: GetAll,
			},
		},
	}
	cmds := []struct {
		Name  string
		Alias string
		Fn    func(ctx context.Context, c *cli.Command) error
	}{
		{"create", "c", Create},
		{"get", "g", Get},
		{"update", "u", Update},
		{"delete", "d", Delete},
	}
	for _, cmd := range cmds {
		rootCmd.Commands = append(rootCmd.Commands, &cli.Command{
			Name:    cmd.Name,
			Aliases: []string{cmd.Alias},
			Before:  validArgs,
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
		Str("task", task).
		Msgf("task created %d", len(tasks)-1)

	return nil
}

func GetAll(ctx context.Context, c *cli.Command) error {
	switch c.String("output") {
	case "plain":
		for i, task := range tasks {
			fmt.Println(i, task)
		}
	case "yaml", "yml":
		content, err := yaml.Marshal(tasks)
		handleErr(err, "marshal yaml")
		fmt.Println(string(content))
	case "json":
		content, err := json.Marshal(tasks)
		handleErr(err, "marshal json")
		fmt.Println(string(content))
	case "table":
		fallthrough
	default:
		tw := table.NewWriter()
		tw.SetColumnConfigs([]table.ColumnConfig{
			{Number: 1, Align: text.AlignLeft},
		})
		tw.SetStyle(table.StyleBold)
		tw.AppendHeader(table.Row{"ID", "Task"})
		for i, task := range tasks {
			if i&1 == 0 {
				tw.AppendRow(table.Row{
					color.BlueString("%d", i),
					color.YellowString(task),
				})
				continue
			}
			tw.AppendRow(table.Row{
				color.GreenString("%d", i),
				color.CyanString(task),
			})
		}
		fmt.Println(tw.Render())
	}
	return nil
}

func Get(ctx context.Context, c *cli.Command) error {
	i, err := strconv.Atoi(c.Args().First())
	handleErr(err, "invalid index")
	handleOutOfLength(i)

	log.Info().Msgf("%d: %s", i, tasks[i])

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
		Str("to", tasks[i]).
		Msgf("update task %d: %s", i, prev)

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
		Str("task", task).
		Msgf("delete task %d", i)

	return nil
}
