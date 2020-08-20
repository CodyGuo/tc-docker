package command

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

func CombinedOutput(s string) ([]byte, error) {
	fields := strings.Fields(strings.TrimSpace(s))
	if len(fields) == 0 {
		return nil, errors.New("Invalid command")
	}
	var c *exec.Cmd
	if len(fields) == 1 {
		c = exec.Command(fields[0])
	} else {
		c = exec.Command(fields[0], fields[1:]...)
	}
	c.Env = os.Environ()
	return c.CombinedOutput()
}
