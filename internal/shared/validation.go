package shared

import (
	"strconv"
	"strings"
)

func NormalizeNumericID(id string) (string, bool) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", false
	}

	value, err := strconv.ParseInt(id, 10, 64)
	if err != nil || value <= 0 {
		return "", false
	}

	return strconv.FormatInt(value, 10), true
}

func NormalizePagination(limit, offset int, defaultLimit, maxLimit int) (int, int) {
	if defaultLimit <= 0 {
		defaultLimit = 20
	}
	if maxLimit <= 0 {
		maxLimit = 100
	}
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
