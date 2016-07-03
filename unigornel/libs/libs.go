package libs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v2"

	"github.com/unigornel/unigornel/unigornel/git"
	"github.com/urfave/cli"
)

const (
	libraryFileFlagName = "libs"
	fetchFlagName       = "fetch"
)

func libraryFileFlag() cli.Flag {
	return cli.StringFlag{
		Name:   libraryFileFlagName,
		EnvVar: "UNIGORNEL_LIBRARIES",
		Usage:  "path to the file containing the unigornel libraries",
		Value:  "libraries.yaml",
	}
}

func fetchFlag() cli.Flag {
	return cli.BoolFlag{
		Name:  fetchFlagName,
		Usage: "issue 'git fetch' before checking out the ref",
	}
}

func Libs() cli.Command {
	return cli.Command{
		Name:  "libs",
		Usage: "manage unigornel libraries",
		Flags: []cli.Flag{
			libraryFileFlag(),
		},
		Action: func(ctx *cli.Context) error {
			o := showLibOptions{
				File: ctx.String(libraryFileFlagName),
			}
			if err := o.showLibs(); err != nil {
				return cli.NewExitError("error: "+err.Error(), 1)
			}
			return nil
		},
		Subcommands: []cli.Command{
			{
				Name:  "save",
				Usage: "save the libraries to a file",
				Action: func(ctx *cli.Context) error {
					o := saveLibOptions{
						File: ctx.GlobalString(libraryFileFlagName),
					}
					if err := o.saveLibs(); err != nil {
						return cli.NewExitError("error: "+err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "update",
				Usage: "update the libraries from a file",
				Flags: []cli.Flag{
					fetchFlag(),
				},
				Action: func(ctx *cli.Context) error {
					o := updateLibOptions{
						File:        ctx.GlobalString(libraryFileFlagName),
						ShouldFetch: ctx.Bool(fetchFlagName),
					}
					if err := o.updateLibs(); err != nil {
						return cli.NewExitError("error: "+err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

type Package struct {
	Name string `yaml:"name"`
	Ref  string `yaml:"ref"`
}

func (lib Package) String() string {
	return fmt.Sprintf("%v (ref: %v)", lib.Name, lib.Ref)
}

type Libraries struct {
	Packages []Package `yaml:"packages"`
}

type showLibOptions struct {
	File string
}

func (o *showLibOptions) showLibs() error {
	libs, err := readLibraries(o.File)
	if err != nil {
		return err
	}

	for _, l := range libs.Packages {
		fmt.Println(l)
	}
	return nil
}

type saveLibOptions struct {
	File string
}

func (o *saveLibOptions) saveLibs() error {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return fmt.Errorf("GOPATH is not set")
	}

	libs, err := readLibraries(o.File)
	if err != nil {
		return err
	}

	curdir, err := os.Getwd()
	if err != nil {
		return err
	}

	var didErr bool
	for i, p := range libs.Packages {
		path := path.Join(gopath, "src", p.Name)
		if err := os.Chdir(path); err != nil {
			continue
		}

		ref, err := git.ShowRef()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not save %v: %v\n", p.Name, err)
			didErr = true
			continue
		}

		if p.Ref != ref {
			fmt.Printf("updating %v: %v -> %v\n", p.Name, p.Ref, ref)
			libs.Packages[i].Ref = ref
		}
	}

	if err := os.Chdir(curdir); err != nil {
		return err
	}

	new, err := yaml.Marshal(&libs)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(o.File, new, 0644); err != nil {
		return err
	}

	if didErr {
		err = fmt.Errorf("could not save some packages")
	}

	return nil
}

type updateLibOptions struct {
	File        string
	ShouldFetch bool
}

func (o *updateLibOptions) updateLibs() error {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return fmt.Errorf("GOPATH is not set")
	}

	libs, err := readLibraries(o.File)
	if err != nil {
		return err
	}

	curdir, err := os.Getwd()
	if err != nil {
		return err
	}

	var didErr bool
	for _, p := range libs.Packages {
		path := path.Join(gopath, "src", p.Name)
		if err := os.Chdir(path); err != nil {
			continue
		}

		if o.ShouldFetch {
			fmt.Printf("fetching %v\n", p.Name)
			git.Fetch()
		}

		fmt.Printf("updating %v to %v\n", p.Name, p.Ref)
		if err := git.Checkout(p.Ref); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not update %v: %v\n", p.Name, err)
			didErr = true
		}
	}

	if err := os.Chdir(curdir); err != nil {
		return err
	}

	if didErr {
		return fmt.Errorf("could not update some packages")
	}

	return nil
}

func readLibraries(file string) (Libraries, error) {
	var libs Libraries
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return libs, err
	}

	if err := yaml.Unmarshal(b, &libs); err != nil {
		return libs, err
	}

	return libs, nil
}
