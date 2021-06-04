package cmdline_service

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestPipe(t *testing.T) {

	savedStdout := os.Stdout // keep backup of the real stdout
	savedStderr := os.Stderr

	reader, writer, _ := os.Pipe()

	done := make(chan bool)

	go func() {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		defer reader.Close()

		shellCommand := `awk '{ print $2 " - " $1 }'`
		cmd := exec.Command(ShellToUse, "-c", shellCommand)

		cmd.Stdin = reader
		cmd.Stdout = &stdout
		cmd.Stdout = &stderr

		if err := cmd.Run(); err != nil {
			fmt.Println(err)
		} else {
			fmt.Fprintln(savedStdout, stdout.String())
			fmt.Fprintln(savedStderr, stderr.String())
		}
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		done <- true
	}()

	os.Stdout = writer

	fmt.Println("a man a plan")
	fmt.Println("via golang")
	writer.Close()
	<-done
	os.Stdout = savedStdout
	fmt.Println("done")
}
