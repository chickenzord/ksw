package main

import "fmt"

var (
	logPrefix = "ksw"
)

func logf(format string, a ...any) {
	fmt.Printf("%s: %s\n", logPrefix, fmt.Sprintf(format, a...))
}
