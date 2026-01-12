package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mergewayhq/mergeway-cli/internal/config"
	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold repository structure",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := contextFromCommand(cmd)
			if err != nil {
				return err
			}

			if len(args) > 0 {
				_, _ = fmt.Fprintln(ctx.Stderr, "init: no arguments are supported")
				return newExitError(1)
			}

			configPath := ctx.Config
			if err := ensureFile(configPath, defaultConfigTemplate()); err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "init: %v\n", err)
				return newExitError(1)
			}

			_, _ = fmt.Fprintf(ctx.Stdout, "Initialized repository at %s\n", ctx.Root)
			return nil
		},
	}

	return cmd
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
	return fmt.Sprintf(`# This configuration file describes your mergeway set-up.

mergeway:
  version: %[1]d  # Current config version. Leave at %[1]d unless release notes say otherwise.

# Everything below lives under the entities map. Uncomment the sample and tweak names
# to match your domain. You can duplicate the block to describe additional entities.
entities:
  # User:
  #   description: Describe what this entity stores.
  #   fields:
  #     id:
  #       type: string
  #       required: true        # Validation fails if required fields are missing.
  #     name: string            # Shorthand for {type: string, required: false}.
  #     email:
  #       type: string
  #       format: email         # Try formats, enums, references, or custom scalars.
  #       description: Shown in notification emails.
  #   identifier: id            # Define which field to use as the identifier of the entity.
  #   include:
  #     - data/users/*.yaml     # Globs for YAML/JSON files on disk.
  #       selector: $.users[*]  # Optional JSONPath when pulling multiple records per file.
  #   data:
  #     - id: user-0001         # You can inline object, this works well for small data sets.
  #       name: Jane Doe
  #       email: jane@example.com
`, config.CurrentVersion)
}
