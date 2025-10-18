package main

import (
	"fmt"
	"os"

	"github.com/riywo/loginshell"
	"github.com/urfave/cli/v2"
)

var (
	version = "dev"
)

func main() {
	app := &cli.App{
		Name:            "ksw",
		Usage:           "kubeconfig switcher",
		Description:     "start a new shell with specified kube context",
		Action:          appAction,
		ArgsUsage:       "[context-name]",
		HideHelpCommand: true,
		Version:         version,
		HideVersion:     false,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func appAction(c *cli.Context) error {
	var contextName string

	if c.Args().Len() == 0 {
		c, err := findContext()
		if err != nil {
			return err
		}

		contextName = c
	} else {
		contextName = c.Args().First()
	}

	shell, err := loginshell.Shell()
	if err != nil {
		return err
	}

	// If already in a ksw session, switch context in-place instead of nesting
	if os.Getenv("KSW_KUBECONFIG_ORIGINAL") != "" {
		return switchContext(contextName)
	}

	return startShell(shell, contextName)
}
