package tests

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

func Build(w io.Writer, name, pack string) (string, error) {
	fh, err := ioutil.TempFile("", "unigornel-tests-")
	if err != nil {
		return "", err
	}
	fh.Close()
	file := fh.Name()

	fmt.Fprintf(w, "[+] building %s to %s\n", name, file)
	cmd := exec.Command(
		"unigornel",
		"build",
		"-x", "-a",
		"-o", file,
		pack,
	)
	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Run(); err != nil {
		os.Remove(file)
		return "", err
	}
	return file, nil
}
