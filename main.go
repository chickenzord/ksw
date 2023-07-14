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

	if len(os.Args) == 1 {
		fmt.Printf("usage: %s context-name\n", os.Args[0])
		return
	}

	contextName := os.Args[1]
	if err := startShell(shell, contextName); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
