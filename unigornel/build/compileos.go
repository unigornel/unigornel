package build

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/urfave/cli"
	"github.ugent.be/unigornel/unigornel/env"
	"github.ugent.be/unigornel/unigornel/exec"
)

// CompileOS is the `compile-os` command
func CompileOS() cli.Command {
	return cli.Command{
		Name:      "compile-os",
		Usage:     "compile Mini-OS with a Go c-archive",
		ArgsUsage: "C-ARCHIVE",
		Flags: []cli.Flag{
			outputFlag(),
		},
		Action: func(ctx *cli.Context) error {
			options := OSOptions{
				Output: ctx.String(outputFlagName),
			}
			if ctx.NArg() != 1 {
				cli.ShowSubcommandHelp(ctx)
				return cli.NewExitError("error: mssing required argument: c-archive", 1)
			}
			options.CArchive = ctx.Args()[0]

			minios, err := env.RequireMiniOSRoot()
			if err != nil {
				return err
			}
			options.MiniOSRoot = minios

			if err := compileOS(options); err != nil {
				return cli.NewExitError("error: "+err.Error(), 1)
			}
			return nil
		},
	}
}

type OSOptions struct {
	MiniOSRoot string
	CArchive   string
	Output     string
}

func compileMiniOSWithCArchive(options OSOptions) error {
	archive, err := filepath.Abs(options.CArchive)
	if err != nil {
		return err
	}

	fmt.Println("[+] compiling mini-os with", options.CArchive)
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(dir)

	if err := os.Chdir(options.MiniOSRoot); err != nil {
		return err
	}

	return exec.InTerminal("make", "GOARCHIVE="+archive).Run()
}

func copyUnikernel(options OSOptions) error {
	if options.Output != "" {
		fmt.Println("[+] copying unikernel to", options.Output)
		return exec.InTerminal("cp", path.Join(options.MiniOSRoot, "mini-os"), options.Output).Run()
	} else {
		fmt.Println("[+] your unikernel is in the minios tree")
		return nil
	}
}

func compileOS(options OSOptions) error {
	if err := compileMiniOSWithCArchive(options); err != nil {
		return err
	}

	return copyUnikernel(options)
}
