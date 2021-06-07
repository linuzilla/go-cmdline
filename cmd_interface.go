package cmdline_service

type CommandInterface interface {
	Command() string
	Execute(args ...string) int
}
