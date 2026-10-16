// Package okf parses Open Knowledge Format concept documents.
package okf

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// IsFile reports whether path has an extension used by OKF concepts.
func IsFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".okf":
		return true
	default:
		return false
	}
}

// Parse extracts the YAML frontmatter and Markdown body from an OKF concept.
func Parse(content []byte) (map[string]any, string, error) {
	text := string(content)
	frontmatterStart, ok := delimiterEnd(text, 0)
	if !ok {
		return nil, "", fmt.Errorf("OKF document must begin with a YAML frontmatter delimiter")
	}

	frontmatterEnd, ok := nextDelimiter(text, frontmatterStart)
	if !ok {
		return nil, "", fmt.Errorf("OKF document has an unclosed YAML frontmatter block")
	}

	frontmatter := make(map[string]any)
	if err := yaml.Unmarshal([]byte(text[frontmatterStart:frontmatterEnd.start]), &frontmatter); err != nil {
		return nil, "", fmt.Errorf("parse OKF frontmatter: %w", err)
	}
	return frontmatter, text[frontmatterEnd.end:], nil
}

type delimiter struct {
	start int
	end   int
}

func nextDelimiter(text string, offset int) (delimiter, bool) {
	for offset <= len(text) {
		end, ok := delimiterEnd(text, offset)
		if ok {
			return delimiter{start: offset, end: end}, true
		}
		next := strings.IndexByte(text[offset:], '\n')
		if next < 0 {
			break
		}
		offset += next + 1
	}
	return delimiter{}, false
}

func delimiterEnd(text string, start int) (int, bool) {
	if start > len(text) {
		return 0, false
	}
	lineEnd := strings.IndexByte(text[start:], '\n')
	if lineEnd < 0 {
		lineEnd = len(text)
	} else {
		lineEnd += start
	}
	line := strings.TrimSuffix(text[start:lineEnd], "\r")
	if line != "---" {
		return 0, false
	}
	if lineEnd == len(text) {
		return lineEnd, true
	}
	return lineEnd + 1, true
}
