package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func cmdInit(ctx *Context, args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(ctx.Stderr)

	if err := fs.Parse(args); err != nil {
		return 1
	}

	configPath := ctx.Config
	if err := ensureFile(configPath, defaultConfigTemplate()); err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "init: %v\n", err)
		return 1
	}

	_, _ = fmt.Fprintf(ctx.Stdout, "Initialized repository at %s\n", ctx.Root)
	return 0
}

func ensureFile(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(content), 0o644)
}

func defaultConfigTemplate() string {
	return `# This configuration file describes your mergeway set-up.

version: 1  # Current config version. Leave at 1 unless release notes say otherwise.

# Everything below lives under the entities map. Uncomment the sample and tweak names
# to match your domain. You can duplicate the block to describe additional entities.
entities:
  # User:
  #   fields:
  #     id:
  #       type: string
  #       required: true        # Validation fails if required fields are missing.
  #     name: string            # Shorthand for {type: string, required: false}.
  #     email:
  #       type: string
  #       format: email         # Try formats, enums, references, or custom scalars.
  #   identifier: id            # Define which field to use as the identifier of the entity.
  #   include:
  #     - data/users/*.yaml     # Globs for YAML/JSON files on disk.
  #       selector: $.users[*]  # Optional JSONPath when pulling multiple records per file.
  #   data:
  #     - id: user-0001         # You can inline object, this works well for small data sets.
  #       name: Jane Doe
  #       email: jane@example.com
`
}
