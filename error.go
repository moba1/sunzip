package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
)

var Stderr io.Writer = os.Stderr

type Level string
const (
	fatal Level = "fatal"
	warn Level = "warning"
)

func printToStderr(lv Level, format string, reason error, a ...interface{}) {
	w := bufio.NewWriter(Stderr)
	fmt.Fprintf(w, string(lv)+": "+format+" (reason: %s)\n", append(a, strconv.Quote(reason.Error()))...)
	w.Flush()
}

func Warn(format string, reason error, a ...interface{}) {
	printToStderr(warn, format, reason, a...)
}

func Fatal(format string, reason error, a ...interface{}) {
	printToStderr(fatal, format, reason, a...)
	os.Exit(1)
}
