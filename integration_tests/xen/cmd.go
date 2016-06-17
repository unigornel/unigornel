package xen

import (
	"os/exec"
	"strconv"
)

func Xl(args ...string) *exec.Cmd {
	return exec.Command("xl", args...)
}

func Create(paused bool, config string) *exec.Cmd {
	args := []string{"create", "-f", config}
	if paused {
		args = append(args, "-p")
	}
	return Xl(args...)
}

func Unpause(id int) *exec.Cmd {
	return Xl("unpause", strconv.Itoa(id))
}

func Console(id int) *exec.Cmd {
	return Xl("console", strconv.Itoa(id))
}

func List() *exec.Cmd {
	return Xl("list")
}

func Destroy(id int) *exec.Cmd {
	return Xl("destroy", strconv.Itoa(id))
}
