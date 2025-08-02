package utils

import (
	"fmt"
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
func SplitQueryInt(query string) ([]int, error) {
	var params []int
	splittedQuery := strings.Split(query, "|")

	for i, param := range splittedQuery {
		paramInt, err := strconv.Atoi(param)
		if err != nil {
			return nil, fmt.Errorf("failed to convert parameter at %d: %w", i, err)
		}
		params = append(params, paramInt)
	}

	return params, nil
}
