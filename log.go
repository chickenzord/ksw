package main

import (
	"fmt"
	"os"
)

var (
	logPrefix = "ksw"
)

func logf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", logPrefix, fmt.Sprintf(format, a...))
}
