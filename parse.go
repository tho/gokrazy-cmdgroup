package cmdgroup

import (
	"fmt"
	"strconv"
	"strings"
)

// parseArgs splits the provided arguments into separate command instances
// based on "--" delimiters.
func parseArgs(args []string) []*Instance {
	var instances []*Instance
	var start, end int

	for end < len(args) {
		if args[end] == "--" {
			instances = append(instances, &Instance{
				Args: args[start:end],
			})
			start = end + 1
		}
		end++
	}

	instances = append(instances, &Instance{
		Args: args[start:end],
	})

	return instances
}

// parseInts converts watch specification to instance indexes.
// Accepts "none", "all", or comma-separated indexes.
func parseInts(watch string, maxValue int) ([]int, error) {
	ints := make([]int, 0)

	if watch == "" || watch == "none" {
		return ints, nil
	}

	if watch == "all" {
		for val := range maxValue {
			ints = append(ints, val)
		}
		return ints, nil
	}

	for indexStr := range strings.SplitSeq(watch, ",") {
		indexStr = strings.TrimSpace(indexStr)
		if indexStr == "" {
			continue
		}

		index, err := strconv.Atoi(indexStr)
		if err != nil {
			return nil, fmt.Errorf("parse int: %w", err)
		}

		if index < 0 || index >= maxValue {
			return nil, fmt.Errorf("index out of range: %d", index)
		}

		ints = append(ints, index)
	}

	return ints, nil
}
