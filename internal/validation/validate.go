package validation

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

// Validate runs the requested validation phases for the given repository root and configuration.
func Validate(root string, cfg *config.Config, opts Options) (*Result, error) {
	if cfg == nil {
		return nil, errors.New("validation: config is required")
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("validation: resolve root: %w", err)
	}

	phaseSet := normalizePhases(opts.Phases)
	if phaseSet[PhaseReferences] {
		phaseSet[PhaseSchema] = true
	}

	res := &Result{}

	rawObjects, formatErrs := collectObjects(absRoot, cfg)
	if len(formatErrs) > 0 {
		if opts.FailFast && len(formatErrs) > 1 {
			formatErrs = formatErrs[:1]
		}
		res.Errors = appendFiltered(res.Errors, formatErrs, phaseSet, PhaseFormat)
		return res, nil
	}

	index, schemaErrs := validateSchema(rawObjects, cfg)
	if len(schemaErrs) > 0 {
		if opts.FailFast && len(schemaErrs) > 1 {
			schemaErrs = schemaErrs[:1]
		}
		res.Errors = appendFiltered(res.Errors, schemaErrs, phaseSet, PhaseSchema)
		return res, nil
	}

	if phaseSet[PhaseReferences] {
		referenceErrs := validateReferences(rawObjects, index, cfg)
		res.Errors = append(res.Errors, referenceErrs...)
		if opts.FailFast && len(referenceErrs) > 0 {
			if len(referenceErrs) > 1 {
				res.Errors = res.Errors[:len(res.Errors)-len(referenceErrs)+1]
			}
			return res, nil
		}
	}

	return res, nil
}
