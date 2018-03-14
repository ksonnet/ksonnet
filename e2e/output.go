package e2e

import (
	"bytes"
	"os/exec"
	"syscall"
)

type output struct {
	stdout   string
	stderr   string
	exitCode int
	args     []string
	cmdName  string
}

func runWithOutput(cmd *exec.Cmd) (*output, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var exitCode int
	if err := cmd.Wait(); err != nil {
		switch t := err.(type) {
		default:
			return nil, err
		case *exec.ExitError:
			status, ok := t.Sys().(syscall.WaitStatus)
			if !ok {
				return nil, t
			}
			exitCode = status.ExitStatus()
		}
	}

	o := &output{
		stdout:   stdout.String(),
		stderr:   stderr.String(),
		exitCode: exitCode,
		args:     cmd.Args,
		cmdName:  cmd.Path,
	}

	return o, nil
}
