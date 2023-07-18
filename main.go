package main

import (
	"fmt"
	"os"

	"github.com/riywo/loginshell"
)

func main() {
	shell, err := loginshell.Shell()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var contextName string

	if len(os.Args) == 1 {
		c, err := pickContext()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		contextName = c
	} else {
		contextName = os.Args[1]
	}

	if err := startShell(shell, contextName); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
