package env

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	"github.ugent.be/unigornel/unigornel/config"

	"github.com/urfave/cli"
)

const (
	MiniOSRootEnv = "UNIGORNEL_MINIOS"
)

const (
	configFullFlagName = "config"
)

func configFlag() cli.Flag {
	return cli.StringFlag{
		Name:   configFullFlagName + ", c",
		EnvVar: "UNIGORNEL_CONFIG",
		Usage:  "path to the configuration file (yaml)",
	}
}

func Env() cli.Command {
	return cli.Command{
		Name:  "env",
		Usage: "setup your shell",
		Flags: []cli.Flag{
			configFlag(),
		},
		Action: func(ctx *cli.Context) error {
			options := EnvOptions{
				ConfigFile: ctx.String(configFullFlagName),
			}

			if err := env(options); err != nil {
				return cli.NewExitError("error: "+err.Error(), 1)
			}
			return nil
		},
	}
}

func RequireMiniOSRoot() (s string, err error) {
	s = os.Getenv(MiniOSRootEnv)
	if s == "" {
		err = cli.NewExitError("error: environment variable UNIGORNEL_MINIOS not set", 1)
	}
	return
}

func GetConfig(configPath string) (config.Config, error) {
	empty := config.Config{}

	s := configPath
	if s == "" {
		u, err := user.Current()
		if err != nil {
			return empty, err
		}
		s = path.Join(u.HomeDir, ".unigornel.yaml")
	}

	data, err := ioutil.ReadFile(s)
	if err != nil {
		return empty, err
	}

	c, err := config.ParseConfig(data)
	if err != nil {
		return empty, err
	}
	return c, nil
}

type EnvOptions struct {
	ConfigFile string
}

func env(options EnvOptions) error {
	config, err := GetConfig(options.ConfigFile)
	if err != nil {
		return err
	}
	for _, v := range configToEnvVars(config) {
		fmt.Println(export(v))
	}
	return nil
}

type envVar struct {
	Name  string
	Value string
}

func export(v envVar) string {
	return "export " + v.Name + "=\"" + v.Value + "\""
}

func configToEnvVars(c config.Config) []envVar {
	return []envVar{
		{
			Name:  "GOROOT",
			Value: c.GoRoot,
		},
		{
			Name:  MiniOSRootEnv,
			Value: c.MiniOS,
		},
		{
			Name:  "PATH",
			Value: c.GoRoot + "/bin:$PATH",
		},
	}
}
