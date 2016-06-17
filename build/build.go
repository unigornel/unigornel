package build

import (
	"io/ioutil"
	"os"

	"github.com/urfave/cli"
	"github.ugent.be/unigornel/env"
)

// Build is the `build` command.
func Build() cli.Command {
	return cli.Command{
		Name:  "build",
		Usage: "build a unikernel",
		Flags: []cli.Flag{
			buildAllFlag(),
			buildVerboseFlag(),
			outputFlag(),
		},
		Action: func(ctx *cli.Context) error {
			options := BuildOptions{
				Go: GoOptions{
					BuildAll:     ctx.Bool(buildAllFlagName),
					BuildVerbose: ctx.Bool(buildVerboseFlagName),
				},
				OS: OSOptions{
					Output: ctx.String(outputFlagName),
				},
			}

			if ctx.NArg() > 1 {
				cli.ShowSubcommandHelp(ctx)
				return cli.NewExitError("error: subcommand expects zero or one arguments", 1)
			} else if ctx.NArg() == 1 {
				options.Go.Package = ctx.Args()[0]
			}

			minios, err := env.RequireMiniOSRoot()
			if err != nil {
				return err
			}
			options.Go.MiniOSRoot = minios
			options.OS.MiniOSRoot = minios

			if err := options.buildAll(); err != nil {
				return cli.NewExitError("error: "+err.Error(), 1)
			}
			return nil
		},
	}
}

type BuildOptions struct {
	Go GoOptions
	OS OSOptions
}

func (o *BuildOptions) buildTemporaryCArchive() error {
	fh, err := ioutil.TempFile("", "unigornel-carchive")
	if err != nil {
		return err
	}
	f := fh.Name()
	o.Go.Output = f
	o.OS.CArchive = f

	if err := compileGo(o.Go); err != nil {
		os.Remove(fh.Name())
		return err
	}
	return nil
}

func (o *BuildOptions) buildAll() error {
	if err := o.buildTemporaryCArchive(); err != nil {
		return err
	}
	defer os.Remove(o.Go.Output)

	return compileOS(o.OS)
}
