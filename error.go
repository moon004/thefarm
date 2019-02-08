package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
)

// Errs handles the error for TheFarm
func Errs(err error) {
	if err != nil {
		red := color.New(color.FgRed).SprintFunc()
		log.Fatalf(fmt.Sprintf("%s %v", red("error:"), err))
		os.Exit(1) // exit with error code 1
	}
}
