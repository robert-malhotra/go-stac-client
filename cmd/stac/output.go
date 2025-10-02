package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"iter"
)

func collectForCLI[T any](seq iter.Seq2[*T, error], marshal func(*T) ([]byte, error)) ([][]byte, error) {
	var (
		results [][]byte
		iterErr error
	)

	seq(func(value *T, err error) bool {
		if err != nil {
			iterErr = err
			return false
		}
		data, err := marshal(value)
		if err != nil {
			iterErr = err
			return false
		}
		results = append(results, data)
		return true
	})

	if iterErr != nil {
		return nil, iterErr
	}
	return results, nil
}

func printJSONArray(entries [][]byte) error {
	if _, err := fmt.Fprintln(os.Stdout, "["); err != nil {
		return err
	}
	for i, entry := range entries {
		if i > 0 {
			if _, err := fmt.Fprintln(os.Stdout, ","); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(os.Stdout, string(entry)); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(os.Stdout, "]")
	return err
}

const interactivePageSize = 10

func printJSONArrayInteractive[T any](seq iter.Seq2[*T, error], marshal func(*T) ([]byte, error)) error {
	if _, err := fmt.Fprintln(os.Stdout, "["); err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)
	var (
		printedAny bool
		processed  int
		iterErr    error
	)

	abort := false
	seq(func(value *T, err error) bool {
		if err != nil {
			iterErr = err
			return false
		}

		data, err := marshal(value)
		if err != nil {
			iterErr = err
			return false
		}

		if printedAny {
			if _, err := fmt.Fprintln(os.Stdout, ","); err != nil {
				iterErr = err
				return false
			}
		}

		if _, err := fmt.Fprintln(os.Stdout, string(data)); err != nil {
			iterErr = err
			return false
		}

		printedAny = true
		processed++

		if interactivePageSize > 0 && processed%interactivePageSize == 0 {
			if _, err := fmt.Fprint(os.Stderr, "Press Enter to continue, or type 'q' to quit: "); err != nil {
				iterErr = err
				return false
			}

			input, err := reader.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					return true
				}
				iterErr = err
				return false
			}

			if strings.EqualFold(strings.TrimSpace(input), "q") {
				abort = true
				return false
			}
		}

		return true
	})

	if _, err := fmt.Fprintln(os.Stdout, "]"); err != nil && iterErr == nil {
		iterErr = err
	}

	if iterErr != nil {
		return iterErr
	}

	if abort {
		return nil
	}

	return nil
}
