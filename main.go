package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

type Input interface {
	Open() (io.ReadCloser, error)
}

type FileInput struct {
	Path string
}

func (f FileInput) Open() (io.ReadCloser, error) {
	file, err := os.Open(f.Path)
	if err != nil {
		return nil, err
	}
	return file, nil
}

type StdinInput struct{}

func (s StdinInput) Open() (io.ReadCloser, error) {
	return io.NopCloser(os.Stdin), nil
}

type CopyFunc func(dst io.Writer, src io.Reader) error

func copyPlain(dst io.Writer, src io.Reader) error {
	_, err := io.Copy(dst, src)
	return err
}

func copyWithLineNumbers(dst io.Writer, src io.Reader) error {
	reader := bufio.NewReader(src)
	number := 1
	for {
		line, readErr := reader.ReadString('\n')

		if len(line) > 0 {
			fmt.Fprintf(dst, "%6d\t%s", number, line)
		}

		number++

		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				return nil
			}
			return readErr
		}
	}
}

func copyWithNonBlankLineNumbers(dst io.Writer, src io.Reader) error {
	reader := bufio.NewReader(src)
	number := 1
	for {
		line, readErr := reader.ReadString('\n')

		if len(line) > 0 {
			if line == "\n" {
				fmt.Fprint(dst, line)
			} else {
				fmt.Fprintf(dst, "%6d\t%s", number, line)
				number++
			}
		}

		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				return nil
			}
			return readErr
		}
	}
}

func chooseCopyFn(nFlag, bFlag bool) CopyFunc {
	if bFlag {
		return copyWithNonBlankLineNumbers
	}
	if nFlag {
		return copyWithLineNumbers
	}
	return copyPlain
}

func buildInputs(args []string) []Input {
	if len(args) == 0 {
		return []Input{StdinInput{}}
	}

	inputs := make([]Input, 0, len(args))
	for _, arg := range args {
		if arg == "-" {
			inputs = append(inputs, StdinInput{})
		} else {
			inputs = append(inputs, FileInput{Path: arg})
		}
	}
	return inputs
}

func processInput(input Input, copyFn CopyFunc) error {
	reader, err := input.Open()
	if err != nil {
		return err
	}
	defer reader.Close()

	return copyFn(os.Stdout, reader)
}

func main() {
	var nFlag, bFlag bool
	flag.BoolVar(&nFlag, "n", false, "number all output lines")
	flag.BoolVar(&bFlag, "b", false, "number non-empty output lines")
	flag.Parse()

	copyFn := chooseCopyFn(nFlag, bFlag)
	inputs := buildInputs(flag.Args())

	for _, input := range inputs {
		if err := processInput(input, copyFn); err != nil {
			fmt.Fprintf(os.Stderr, "cccat: %s\n", err)
		}
	}
}
