package build

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.ugent.be/unigornel/env"
	"github.ugent.be/unigornel/exec"

	"github.com/urfave/cli"
)

const (
	buildAllFlagName     = "a"
	buildVerboseFlagName = "x"
	outputFlagName       = "o"
)

func buildAllFlag() cli.Flag {
	return cli.BoolFlag{
		Name:  buildAllFlagName,
		Usage: "recompile everything (corresponds to go's -a flag)",
	}
}

func buildVerboseFlag() cli.Flag {
	return cli.BoolFlag{
		Name:  buildVerboseFlagName,
		Usage: "be verbose (corresponds to go's -x flag)",
	}
}

func outputFlag() cli.Flag {
	return cli.StringFlag{
		Name:  outputFlagName,
		Usage: "output file (c-archive or unikernel)",
	}
}

// CompileGo is the `compile-go` command
func CompileGo() cli.Command {
	return cli.Command{
		Name:      "compile-go",
		Usage:     "compile a Go application to an intermediate c-archive",
		ArgsUsage: "[PACKAGE]",
		Flags: []cli.Flag{
			buildAllFlag(),
			buildVerboseFlag(),
			outputFlag(),
		},
		Action: func(ctx *cli.Context) error {
			options := GoOptions{
				BuildAll:     ctx.Bool(buildAllFlagName),
				BuildVerbose: ctx.Bool(buildVerboseFlagName),
				Output:       ctx.String(outputFlagName),
			}
			if options.Output == "" {
				cli.ShowSubcommandHelp(ctx)
				return cli.NewExitError("error: missing required flag -o", 1)
			}

			if ctx.NArg() > 1 {
				cli.ShowSubcommandHelp(ctx)
				return cli.NewExitError("error: subcommand expects zero or one arguments", 1)
			} else if ctx.NArg() == 1 {
				options.Package = ctx.Args()[0]
			}

			minios, err := env.RequireMiniOSRoot()
			if err != nil {
				return err
			}
			options.MiniOSRoot = minios

			if err := compileGo(options); err != nil {
				return cli.NewExitError("error: "+err.Error(), 1)
			}
			return nil
		},
	}
}

type GoOptions struct {
	BuildAll     bool
	BuildVerbose bool
	Package      string
	MiniOSRoot   string
	Output       string
}

func generateMiniOSLinks(options GoOptions) error {
	fmt.Println("[+] preparing mini-os")
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(dir)

	if err := os.Chdir(options.MiniOSRoot); err != nil {
		return err
	}

	return exec.InTerminal("make", "links").Run()
}

func compileCArchive(options GoOptions) error {
	fmt.Printf("[+] compiling Go to a c-archive (%s)\n", options.Output)
	include := strings.Join([]string{
		"-isystem", path.Join(options.MiniOSRoot, "include"),
		"-isystem", path.Join(options.MiniOSRoot, "include", "x86"),
		"-isystem", path.Join(options.MiniOSRoot, "include", "x86", "x86_64"),
	}, " ")

	envs := append(os.Environ(), []string{
		"CGO_ENABLED=1",
		"CGO_CFLAGS=" + include,
		"GOOS=unigornel",
		"GOARCH=amd64",
	}...)

	args := []string{"build", "-buildmode=c-archive"}
	args = append(args, "-o", options.Output)

	if options.BuildAll {
		args = append(args, "-a")
	}
	if options.BuildVerbose {
		args = append(args, "-x")
	}

	if options.Package != "" {
		args = append(args, options.Package)
	}

	defer func() {
		f := options.Output
		p := f[:len(f)-len(path.Ext(f))] + ".h"
		fmt.Println("[+] removing:", p)
		if err := os.Remove(p); err != nil {
			fmt.Println("[-] warning:", err)
		}
	}()
	cmd := exec.InTerminal("go", args...)
	cmd.Env = envs
	return cmd.Run()
}

func fixCArchive(options GoOptions) error {
	fmt.Println("[+] fixing up c-archive for mini-os")
	return exec.InTerminal(
		"objcopy",
		"--globalize-symbol=_rt0_amd64_unigornel_lib",
		options.Output,
	).Run()
}

func compileGo(options GoOptions) error {
	if err := generateMiniOSLinks(options); err != nil {
		return err
	}
	if err := compileCArchive(options); err != nil {
		return err
	}
	if err := fixCArchive(options); err != nil {
		return err
	}

	fmt.Printf("[+] c-archive is in '%s'\n", options.Output)
	return nil
}
