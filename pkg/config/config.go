package config

import internalconfig "github.com/mergewayhq/mergeway-cli/internal/config"

// Config re-exports the internal configuration struct for external consumers.
type Config = internalconfig.Config

// TypeDefinition exposes entity type metadata.
type TypeDefinition = internalconfig.TypeDefinition

// FieldDefinition exposes field metadata for a type.
type FieldDefinition = internalconfig.FieldDefinition

// IdentifierDefinition exposes identifier metadata for a type.
type IdentifierDefinition = internalconfig.IdentifierDefinition

// IncludeDefinition exposes include directives for a type.
type IncludeDefinition = internalconfig.IncludeDefinition

// WriteDefaults exposes global write descriptors.
type WriteDefaults = internalconfig.WriteDefaults

// WriteDefinition exposes per-type write configuration.
type WriteDefinition = internalconfig.WriteDefinition

// WriteFormat enumerates supported output formats.
type WriteFormat = internalconfig.WriteFormat

// Load reads the mergeway configuration and resolves includes.
var Load = internalconfig.Load

// CurrentVersion defines the supported mergeway.yaml schema version.
const CurrentVersion = internalconfig.CurrentVersion

const (
	// WriteFormatYAML encodes YAML documents.
	WriteFormatYAML = internalconfig.WriteFormatYAML
	// WriteFormatJSON encodes JSON documents.
	WriteFormatJSON = internalconfig.WriteFormatJSON
)

// DefaultWriteTemplate describes the default file template.
const DefaultWriteTemplate = internalconfig.DefaultWriteTemplate
