package main

import (
	"fmt"
	"os"

	"github.com/riywo/loginshell"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:            "ksw",
		Usage:           "kubeconfig switcher",
		Description:     "start a new shell with specified kube context",
		Action:          mainAction,
		ArgsUsage:       "[context-query]",
		HideHelpCommand: true,
		Version:         Version,
		HideVersion:     false,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "list available contexts without starting a shell",
			},
			&cli.BoolFlag{
				Name:    "env",
				Aliases: []string{"e"},
				Usage:   "print ksw environment variables",
			},
		},
	}

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println(GetVersionInfo().String())
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func mainAction(c *cli.Context) error {
	// Handle --list flag
	if c.Bool("list") {
		return listContextsAction()
	}

	// Handle --env flag
	if c.Bool("env") {
		return envAction()
	}

	// Get initial query from args, or empty string if no args
	query := ""
	if c.Args().Len() > 0 {
		query = c.Args().First()
	}

	// Show fuzzy finder with initial query
	contextName, err := findContext(query)
	if err != nil {
		return err
	}

	// If already in a ksw session, switch context in-place instead of nesting
	if os.Getenv("KSW_KUBECONFIG_ORIGINAL") != "" {
		return switchContext(contextName)
	}

	// Otherwise, start a new shell with the selected context
	shell, err := loginshell.Shell()
	if err != nil {
		return err
	}

	return startShell(shell, contextName)
}

func listContextsAction() error {
	kubeconfigPath := getOriginalKubeconfigPath()

	contexts, err := listContexts(kubeconfigPath)
	if err != nil {
		return err
	}

	for _, ctx := range contexts {
		fmt.Println(ctx)
	}

	return nil
}

func envAction() error {
	envVars := []string{
		"KSW_ACTIVE",
		"KSW_KUBECONFIG",
		"KSW_KUBECONFIG_ORIGINAL",
		"KSW_SHELL",
		"KUBECONFIG",
	}

	for _, envVar := range envVars {
		value := os.Getenv(envVar)
		fmt.Printf("%s=%s\n", envVar, value)
	}

	return nil
}
