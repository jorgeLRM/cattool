package main

import (
	"errors"
	"io"
	"os"
)

type Input interface {
	Open() (io.Reader, error)
}

type FileInput struct {
	Path string
}

func (f FileInput) Open() (io.Reader, error) {
	file, err := os.Open(f.Path)
	if err != nil {
		return nil, errors.New(f.Path + ": No such file or directory")
	}
	return file, nil
}

type StdinInput struct{}

func (s StdinInput) Open() (io.Reader, error) {
	return os.Stdin, nil
}

func resolve(s []string) []Input {
	if len(s) == 0 {
		return []Input{StdinInput{}}
	}

	var inputs []Input
	for _, arg := range s {
		if arg == "-" {
			inputs = append(inputs, StdinInput{})
		} else {
			inputs = append(inputs, FileInput{Path: arg})
		}
	}
	return inputs
}

func main() {
	inputs := resolve(os.Args[1:])

	for _, input := range inputs {
		reader, err := input.Open()
		if err != nil {
			os.Stderr.WriteString("kitten: " + err.Error() + "\n")
			continue
		}

		_, err = io.Copy(os.Stdout, reader)
		if err != nil {
			os.Stderr.WriteString("kitten: " + err.Error() + "\n")
			continue
		}
	}
}
