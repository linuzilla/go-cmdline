package cmdline_service

import (
	"bytes"
	"fmt"
	"github.com/chzyer/readline"
	"github.com/linuzilla/summer"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type cmdlineService struct {
	commandMap map[string]CommandInterface
	prompt     string
}

var regx = regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)`)

const ShellToUse = `/bin/bash`

func New(applicationContext summer.ApplicationContextManager, prompt string) *cmdlineService {
	commandService := &cmdlineService{
		prompt: prompt,
	}
	commandService.commandMap = make(map[string]CommandInterface)

	var cmdInterface CommandInterface

	applicationContext.ForEach(&cmdInterface, func(data interface{}) {
		if cmd, ok := data.(CommandInterface); ok {
			commandService.Register(cmd)
		}
	})
	return commandService
}

func (commandService *cmdlineService) pipeCommand(handler CommandInterface, pipeCmd string, args ...string) {
	savedStdout := os.Stdout // keep backup of the real stdout
	savedStderr := os.Stderr

	reader, writer, _ := os.Pipe()

	done := make(chan bool)

	go func() {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		defer reader.Close()

		cmd := exec.Command(ShellToUse, "-c", pipeCmd)

		cmd.Stdin = reader
		cmd.Stdout = &stdout
		cmd.Stdout = &stderr

		if err := cmd.Run(); err != nil {
			fmt.Println(err)
		} else {
			fmt.Fprintln(savedStdout, stdout.String())
			fmt.Fprintln(savedStderr, stderr.String())
		}
		done <- true
	}()

	os.Stdout = writer

	defer func() {
		os.Stdout = savedStdout
	}()

	handler.Execute(args...)
	writer.Close()
	<-done
}

func (commandService *cmdlineService) RunCommand(command string) {
	if strings.HasPrefix(command, `!`) {
		shellCommand := strings.TrimSpace(command[1:])

		cmd := exec.Command(ShellToUse, "-c", shellCommand)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			fmt.Println(err)
		}
	} else {
		splitArgs := regx.FindAllString(command, -1)
		//fmt.Println("your array: ", splitArgs)
		cmd := splitArgs[0]

		if handler, found := commandService.commandMap[cmd]; found {
			//fmt.Println("Execute: " + cmd)
			args := splitArgs[1:]

			if argsLen := len(args); argsLen > 1 {
				for i, arg := range args {
					if arg == `|` && i < argsLen {
						pipeCmd := args[i+1:]
						commandService.pipeCommand(handler, strings.Join(pipeCmd, ` `), args[:i]...)
						return
					}
				}
			}
			handler.Execute(args...)
		} else {
			fmt.Printf("%s: command not found\n", cmd)
		}

		fmt.Println()
	}
}

func (commandService *cmdlineService) Register(handler CommandInterface) {
	commandService.commandMap[handler.Command()] = handler
}

func (commandService *cmdlineService) Execute() {
	keys := make([]readline.PrefixCompleterInterface, 0, len(commandService.commandMap))
	// regx := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)`)

	fmt.Print("Registered commands:")
	for key := range commandService.commandMap {
		keys = append(keys, readline.PcItem(key))
		fmt.Print(" [" + key + "]")
	}
	fmt.Println()
	fmt.Println()

	var completer = readline.NewPrefixCompleter(keys...)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "\033[34m" + commandService.prompt + ":>\033[0m ",
		AutoComplete: completer,
		EOFPrompt:    "exit",
	})

	if err != nil {
		panic(err)
	}

	defer rl.Close()

	rl.SetVimMode(false)

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		commandService.RunCommand(line)
	}
}
