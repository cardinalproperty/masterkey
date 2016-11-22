package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type Command struct {
	Name   string
	Action ActionFunc
	Usage  string
}

type ActionFunc func([]string) (string, error)

type REPL struct {
	usage    string
	prompt   string
	commands map[string]ActionFunc
	input    io.Reader
	output   io.Writer

	stopchan chan struct{}
}

func New(prompt string) *REPL {
	return &REPL{
		commands: make(map[string]ActionFunc),
		prompt:   prompt,
		stopchan: make(chan struct{}),
		input:    os.Stdin,
		output:   os.Stdout,
	}
}

func (r *REPL) Stop() {
	close(r.stopchan)
}

func (r *REPL) AddCommand(cmd Command) {
	r.commands[cmd.Name] = cmd.Action
	r.usage = r.usage + cmd.Usage + "\n"
}

func (r *REPL) eval(line string) (string, error) {
	args := strings.Split(line, " ")
	command := args[0]

	if command == "help" {
		return r.usage, nil
	}

	action, exists := r.commands[command]
	if !exists {
		return "", fmt.Errorf("command not recognized. Type `help` for a list of commands.")
	}

	res, err := action(args[1:])
	if err != nil {
		return "", err
	}

	return res, nil
}

func (r *REPL) Loop() error {
	msgchan := make(chan string)

	go func() {
		scanner := bufio.NewScanner(r.input)
		for scanner.Scan() {
			msgchan <- scanner.Text()
		}
	}()

	for {
		fmt.Fprint(r.output, r.prompt)
		select {
		case <-r.stopchan:
			return nil
		case line := <-msgchan:
			fmt.Println(line)
			if line != "" {
				res, err := r.eval(line)
				if err != nil {
					fmt.Fprintln(r.output, "error: ", err.Error())
					continue
				}
				fmt.Fprintln(r.output, res)
			}
		}
	}
}
