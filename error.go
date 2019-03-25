package main

import (
	"fmt"
	logs "log" // avoid conflict with log *logger.Logger
	"os"

	"github.com/fatih/color"
)

// Errs handles the error for TheFarm
func Errs(msg string, err error) {
	if err != nil {
		red := color.New(color.FgRed).SprintFunc()
		yellow := color.New(color.FgCyan).SprintFunc()
		logs.Fatalf(fmt.Sprintf("%s %s %s %v",
			yellow("Message:"), msg, red("Error:"), err))
		os.Exit(1) // exit with error code 1
	}
}
