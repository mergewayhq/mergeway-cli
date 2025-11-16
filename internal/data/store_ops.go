package data

import (
	"errors"
	"fmt"
	"sort"
)

// List returns sorted identifiers for the specified type.
func (s *Store) List(typeName string) ([]string, error) {
	typeDef, err := s.requireType(typeName)
	if err != nil {
		return nil, err
	}

	objects, err := s.loadAll(typeDef)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(objects))
	for _, obj := range objects {
		ids = append(ids, obj.ID)
	}

	sort.Strings(ids)
	return ids, nil
}

// Get returns the object by identifier.
func (s *Store) Get(typeName, id string) (*Object, error) {
	if id == "" {
		return nil, errors.New("data: id is required")
	}
	typeDef, err := s.requireType(typeName)
	if err != nil {
		return nil, err
	}

	loc, err := s.findObject(typeDef, id)
	if err != nil {
		return nil, err
	}
	if loc == nil {
		return nil, fmt.Errorf("data: %s %q not found", typeName, id)
	}

	return loc.cloneObject(), nil
}

// LoadAll retrieves all objects of a type.
func (s *Store) LoadAll(typeName string) ([]*Object, error) {
	typeDef, err := s.requireType(typeName)
	if err != nil {
		return nil, err
	}

	objs, err := s.loadAll(typeDef)
	if err != nil {
		return nil, err
	}

	result := make([]*Object, len(objs))
	for i, obj := range objs {
		result[i] = obj.clone()
	}
	return result, nil
}

// Create writes a new object to disk.
func (s *Store) Create(typeName string, fields map[string]any) (*Object, error) {
	typeDef, err := s.requireType(typeName)
	if err != nil {
		return nil, err
	}

	idValue, normalizedID, err := extractIdentifierValue(typeDef, fields)
	if err != nil {
		return nil, fmt.Errorf("data: %s create: %w", typeName, err)
	}

	if existing, _ := s.Get(typeName, idValue); existing != nil {
		return nil, fmt.Errorf("data: %s %q already exists", typeName, idValue)
	}

	target, err := s.chooseCreateTarget(typeDef, idValue)
	if err != nil {
		return nil, err
	}

	doc, err := s.loadFile(target.Path, typeDef.Name, "")
	if err != nil && !errors.Is(err, errFileNotFound) {
		return nil, err
	}

	normalized := cleanFields(fields)
	idField := typeDef.Identifier.Field
	normalized[idField] = normalizedID

	if target.Multi {
		fi := doc
		if fi == nil {
			fi = &fileContent{TypeName: typeDef.Name, Format: target.Format, Multi: true}
		}
		if fi.TypeName == "" {
			fi.TypeName = typeDef.Name
		}
		if fi.TypeName != typeDef.Name {
			return nil, fmt.Errorf("data: file %s declared type %s; expected %s", target.Path, fi.TypeName, typeDef.Name)
		}
		fi.Path = target.Path
		if fi.Items == nil {
			fi.Items = make([]map[string]any, 0)
		}
		fi.Items = append(fi.Items, cloneMap(normalized))
		if err := s.writeFile(target.Path, fi); err != nil {
			return nil, err
		}
		return &Object{
			Type:     typeDef.Name,
			ID:       idValue,
			Fields:   cloneMap(normalized),
			File:     target.Path,
			ReadOnly: false,
		}, nil
	}

	if err := s.writeSingle(target.Path, target.Format, normalized); err != nil {
		return nil, err
	}

	return &Object{
		Type:     typeDef.Name,
		ID:       idValue,
		Fields:   cloneMap(normalized),
		File:     target.Path,
		ReadOnly: false,
	}, nil
}

// Update replaces or merges an object on disk.
func (s *Store) Update(typeName, id string, fields map[string]any, merge bool) (*Object, error) {
	typeDef, err := s.requireType(typeName)
	if err != nil {
		return nil, err
	}

	loc, err := s.findObject(typeDef, id)
	if err != nil {
		return nil, err
	}
	if loc == nil {
		return nil, fmt.Errorf("data: %s %q not found", typeName, id)
	}
	if loc.Inline {
		return nil, fmt.Errorf("data: %s %q is defined inline and cannot be modified", typeName, id)
	}
	if loc.ReadOnly {
		return nil, fmt.Errorf("data: %s %q is sourced via selector include and cannot be modified", typeName, id)
	}

	updated := cloneMap(loc.Object)
	if merge {
		mergeMaps(updated, fields)
	} else {
		updated = cleanFields(fields)
	}

	idField := typeDef.Identifier.Field
	idFieldDef := typeDef.Fields[idField]
	var normalizedID any = id
	if idFieldDef != nil {
		converted, err := coerceIdentifierValue(idFieldDef.Type, idField, id)
		if err != nil {
			return nil, fmt.Errorf("data: %s update: %w", typeName, err)
		}
		normalizedID = converted
	}
	updated[idField] = normalizedID
	removeTypeKeys(updated)

	if loc.Multi {
		loc.File.Items[loc.Index] = cloneMap(updated)
		if err := s.writeFile(loc.FilePath, loc.File); err != nil {
			return nil, err
		}
		return &Object{
			Type:     typeDef.Name,
			ID:       id,
			Fields:   cloneMap(updated),
			File:     loc.FilePath,
			ReadOnly: false,
		}, nil
	}

	if err := s.writeSingle(loc.FilePath, loc.Format, updated); err != nil {
		return nil, err
	}

	return &Object{
		Type:     typeDef.Name,
		ID:       id,
		Fields:   cloneMap(updated),
		File:     loc.FilePath,
		ReadOnly: false,
	}, nil
}

// Delete removes an object from disk.
func (s *Store) Delete(typeName, id string) error {
	typeDef, err := s.requireType(typeName)
	if err != nil {
		return err
	}

	loc, err := s.findObject(typeDef, id)
	if err != nil {
		return err
	}
	if loc == nil {
		return fmt.Errorf("data: %s %q not found", typeName, id)
	}
	if loc.Inline {
		return fmt.Errorf("data: %s %q is defined inline and cannot be modified", typeName, id)
	}
	if loc.ReadOnly {
		return fmt.Errorf("data: %s %q is sourced via selector include and cannot be modified", typeName, id)
	}

	if loc.Multi {
		items := loc.File.Items
		loc.File.Items = append(items[:loc.Index], items[loc.Index+1:]...)
		if len(loc.File.Items) == 0 {
			return s.removeFile(loc.FilePath)
		}
		return s.writeFile(loc.FilePath, loc.File)
	}

	return s.removeFile(loc.FilePath)
}
