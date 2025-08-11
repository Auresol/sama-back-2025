package utils

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Contains is a helper for enum validation.
func Contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

// Seperate each param by |, return as an array of int
func SplitQueryUint(query string) ([]uint, error) {
	var params []uint

	if query == "" {
		return params, nil
	}

	splittedQuery := strings.Split(query, "|")

	for i, param := range splittedQuery {
		paramInt, err := strconv.ParseUint(param, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to convert parameter at %d: %w", i, err)
		}
		params = append(params, uint(paramInt))
	}

	// Sort slices assending
	sort.Slice(params, func(i, j int) bool {
		return params[i] < params[j]
	})

	return params, nil
}
