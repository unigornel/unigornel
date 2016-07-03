package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ShowRef() (string, error) {
	cmd := exec.Command("git", "show-ref", "--head", "-s", "^refs/origin/HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func Checkout(ref string) error {
	out, err := exec.Command("git", "checkout", ref).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v", strings.TrimSpace(string(out)))
	}
	return nil
}

func Fetch(args ...string) error {
	args = append([]string{"fetch"}, args...)
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
