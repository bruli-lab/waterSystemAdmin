package controller

import (
	"errors"
	"slices"
	"strconv"
)

// availableRelays are the relays a zone can be assigned to. The irrigation system has 3.
var availableRelays = []int{1, 2, 3}

var errInvalidRelay = errors.New("invalid relay value")

// parseRelays converts the submitted relay checkboxes and rejects any relay the system doesn't have.
func parseRelays(values []string) ([]int, error) {
	rel := make([]int, 0, len(values))
	for _, val := range values {
		num, err := strconv.Atoi(val)
		if err != nil || !slices.Contains(availableRelays, num) || slices.Contains(rel, num) {
			return nil, errInvalidRelay
		}
		rel = append(rel, num)
	}
	return rel, nil
}
