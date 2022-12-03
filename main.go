package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/call-cc/icfp2006/um"
)

func main() {
	prg, err := ParseArg()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	um.Init(prg)
	um.Spin()
}

func ParseArg() (string, error) {
	if len(os.Args) < 2 {
		return "", errors.New("no argument given on command line")
	}

	return os.Args[1], nil
}
