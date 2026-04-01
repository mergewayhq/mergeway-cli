package diff

import (
	"errors"
	"strings"
)

type diffErrorCategory string

const (
	diffErrorCategoryInput      diffErrorCategory = "input error"
	diffErrorCategoryRepository diffErrorCategory = "repository state error"
	diffErrorCategoryData       diffErrorCategory = "data error"
	diffErrorCategoryInternal   diffErrorCategory = "internal error"
)

func FormatCommandError(err error) string {
	if err == nil {
		return ""
	}

	category := classifyDiffCommandError(err)
	message := normalizeDiffErrorMessage(err.Error())
	return "diff: " + string(category) + ": " + message
}

func classifyDiffCommandError(err error) diffErrorCategory {
	if err == nil {
		return diffErrorCategoryInternal
	}

	if errors.Is(err, ErrTooManyArgs) {
		return diffErrorCategoryInput
	}

	var buildErr *LogicalDatabaseBuildError
	if errors.As(err, &buildErr) {
		return diffErrorCategoryData
	}

	message := err.Error()
	switch {
	case strings.Contains(message, "invalid revision"):
		return diffErrorCategoryInput
	case strings.Contains(message, "resolve revision"):
		if strings.Contains(message, "not a git repository") {
			return diffErrorCategoryRepository
		}
		return diffErrorCategoryInput
	case strings.Contains(message, "not a git repository"):
		return diffErrorCategoryRepository
	case strings.Contains(message, "config file "),
		strings.Contains(message, "mergeway block is required"),
		strings.Contains(message, "mergeway.version"),
		strings.Contains(message, "entity "),
		strings.Contains(message, "include path"),
		strings.Contains(message, "read working tree file"),
		strings.Contains(message, "walk working tree"),
		strings.Contains(message, "git "),
		strings.Contains(message, "inspect "):
		return diffErrorCategoryRepository
	default:
		return diffErrorCategoryInternal
	}
}

func normalizeDiffErrorMessage(message string) string {
	message = strings.TrimSpace(message)
	message = strings.TrimPrefix(message, "diff: ")
	return message
}
