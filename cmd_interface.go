package cmdline_service

type CommandInterface interface {
	ImplementCommandInterface()
	Command() string
	Execute(args ...string) int
}
