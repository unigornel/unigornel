package tests

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

func Build(w io.Writer, name, pack string, other ...string) (string, error) {
	fh, err := ioutil.TempFile("", "unigornel-tests-")
	if err != nil {
		return "", err
	}
	fh.Close()
	file := fh.Name()

	fmt.Fprintf(w, "[+] building %s to %s\n", name, file)
	args := []string{"build", "-x", "-a", "-o", file}
	args = append(args, other...)
	args = append(args, pack)
	cmd := exec.Command(
		"unigornel",
		args...,
	)
	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Run(); err != nil {
		os.Remove(file)
		return "", err
	}
	return file, nil
}

func GoGet(w io.Writer, pack string) error {
	fmt.Fprintf(w, "[+] go get -v -d %v\n", pack)
	cmd := exec.Command("go", "get", "-v", "-d", pack)
	cmd.Stdout = w
	cmd.Stderr = w
	return cmd.Run()
}

func UpdateLibs(w io.Writer) error {
	fmt.Fprintf(w, "[+] unigornel libs update\n")
	cmd := exec.Command("unigornel", "libs", "update")
	cmd.Stdout = w
	cmd.Stderr = w
	return cmd.Run()
}
