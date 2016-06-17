package tests

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"

	"github.ugent.be/unigornel/integration_tests/xen"
)

type SimpleTest struct {
	Name        string
	Category    string
	Package     string
	Memory      int
	Stdin       []byte
	Timeout     time.Duration
	CanCrash    bool
	CanShutdown bool
	CanTimeout  bool
	CheckRun    func(string) error

	unikernel  string
	domain     *xen.Domain
	didTimeout bool
	output     string
}

func SimpleTestPackage(arg ...string) string {
	s := append([]string{"github.ugent.be/unigornel/integration_tests/tests"}, arg...)
	return path.Join(s...)
}

func (t *SimpleTest) GetName() string {
	return t.Name
}

func (t *SimpleTest) GetCategory() string {
	return t.Category
}

func (t *SimpleTest) Build(w io.Writer) error {
	fh, err := ioutil.TempFile("", "unigornel-tests-")
	if err != nil {
		return err
	}
	fh.Close()
	t.unikernel = fh.Name()

	fmt.Fprintf(w, "[+] building %s to %s\n", t.Name, t.unikernel)
	cmd := exec.Command(
		"unigornel",
		"build",
		"-x", "-a",
		"-o", fh.Name(),
		t.Package,
	)
	cmd.Stdout = w
	cmd.Stderr = w

	return cmd.Run()
}

func (t *SimpleTest) Setup(w io.Writer) error {
	kernel := xen.Kernel{
		Binary:  t.unikernel,
		Memory:  t.Memory,
		Name:    t.Name,
		OnCrash: xen.OnCrashPreserve,
	}

	fmt.Fprintln(w, "[+] creating paused kernel")
	dom, err := kernel.CreatePausedUniqueName(func(cmd *exec.Cmd) {
		cmd.Stdout = w
		cmd.Stderr = w
	})
	fmt.Fprintln(w, "[+] domain created:", dom)
	t.domain = dom
	return err
}

func (t *SimpleTest) Run(w io.Writer) error {
	out := bytes.NewBuffer(nil)
	console := t.domain.Console()
	console.Stdout = io.MultiWriter(w, out)
	console.Stderr = console.Stdout
	stdin, err := console.StdinPipe()
	if err != nil {
		return err
	}

	fmt.Fprintln(w, "[+] attching to the console")
	if err := console.Start(); err != nil {
		return err
	}

	if t.Stdin != nil {
		fmt.Fprintln(w, "[+] writing to console")
		if _, err := stdin.Write(t.Stdin); err != nil {
			return err
		}
	}

	done := make(chan struct{})
	exited := make(chan error)
	timeout := make(chan struct{})
	go func() {
		console.Wait()
		close(done)
	}()
	go func() {
		for {
			select {
			case <-done:
				return
			case <-timeout:
				return
			case <-time.After(1 * time.Second):
				c, err := t.domain.Update()
				if err != nil {
					exited <- err
					close(exited)
					return
				}
				if c.State.Check(xen.DomainStateShutdown) || c.State.Check(xen.DomainStateCrashed) {
					close(exited)
					return
				}
			}
		}
	}()
	go func() {
		time.Sleep(t.Timeout)
		close(timeout)
	}()

	fmt.Fprintln(w, "[+] unpausing unikernel domain")
	if err := t.domain.Unpause().Run(); err != nil {
		console.Process.Kill()
		return err
	}

	select {
	case <-done:
		return errors.New("console unexpectedly exited")
	case err := <-exited:
		// Make sure the console can catch up
		time.Sleep(1 * time.Second)
		console.Process.Kill()
		if err != nil {
			return err
		}
		t.didTimeout = false
		<-done
	case <-timeout:
		t.didTimeout = true
		console.Process.Kill()
		<-done
	}
	t.output = string(out.Bytes())

	if t.didTimeout && !t.CanTimeout {
		return errors.New("unikernel timed out")
	}

	domain, err := t.domain.Update()
	if err != nil {
		return errors.New("domain not preserved: " + err.Error())
	}

	if domain.State.Check(xen.DomainStateCrashed) && !t.CanCrash {
		return errors.New("domain crashed")
	}

	if domain.State.Check(xen.DomainStateShutdown) && !t.CanShutdown {
		return errors.New("domain shutdown")
	}
	return nil
}

func (t *SimpleTest) Check(w io.Writer) error {
	if t.CheckRun == nil {
		fmt.Fprintln(w, "[+] no output checks specified")
		return nil
	}
	fmt.Fprintln(w, "[+] checking unikernel output")
	if err := t.CheckRun(t.output); err != nil {
		fmt.Fprintf(w, "[-] check error: %v\n", err)
		return err
	}
	return nil
}

func (t *SimpleTest) Clean(w io.Writer, success bool) (err error) {
	if t.unikernel != "" {
		fmt.Fprintf(w, "[+] removing %s\n", t.unikernel)
		os.Remove(t.unikernel)
	}

	if t.domain != nil {
		cmd := t.domain.Destroy()
		cmd.Stderr = w
		cmd.Stdout = w
		err = cmd.Run()
	}

	return
}
