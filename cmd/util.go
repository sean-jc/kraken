package cmd

import (
	"os"
)

func inputFile(name string) (*os.File, error) {
	if name == "@" {
		return os.Stdin, nil
	}
	return os.Open(name)
}

func outputFile(name string) (*os.File, error) {
	if name == "@" {
		return os.Stdout, nil
	}
	return os.Create(name)
}
