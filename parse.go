package main

import (
	"fmt"
	"iter"
	"slices"
	"strconv"
	"strings"
)

// parseArgs splits arguments into sections on "--" delimiters.
func parseArgs(args []string) [][]string {
	return slices.Collect(slicesSplitSeq(args, "--"))
}

// slicesSplitSeq returns an iterator over sub-slices of s split around the
// separator element. It behaves like [strings.SplitSeq] but operates on slices
// of a comparable type.
func slicesSplitSeq[E comparable](s []E, sep E) iter.Seq[[]E] {
	return func(yield func([]E) bool) {
		start := 0
		for i, v := range s {
			if v == sep {
				if !yield(s[start:i]) {
					return
				}
				start = i + 1
			}
		}
		yield(s[start:])
	}
}

// parseInts parses a comma-separated list of integers.
func parseInts(s string) ([]int, error) {
	var ints []int

	for part := range strings.SplitSeq(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("parse int: %w", err)
		}

		ints = append(ints, n)
	}

	return ints, nil
}
