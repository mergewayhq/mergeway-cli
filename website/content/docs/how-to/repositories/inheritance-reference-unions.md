---
title: Model inheritance and reference unions
description: Reuse native fields through entity inheritance and point one field at multiple entity types.
type: docs
mode: how-to
weight: 30
bookToC: true
---

# Model inheritance and reference unions

Entity inheritance models an assignable family; a reference union models a value that may target one of several independent entities. These procedures target Mergeway v0.10.0 and the release-equivalent `examples/inheritance` and `examples/reference-union` fixtures.

## Reuse a parent schema

In `examples/inheritance`, `Animal` supplies identity and common fields while `Dog` adds `breed`:

```yaml
Animal:
  identifier: id
  fields:
    id: { type: string, required: true }
    name: { type: string, required: true }

Dog:
  extends: Animal
  include:
    - data/dogs/*.yaml
  fields:
    breed: { type: string, required: true }
```

`Animal` may remain schema-only because it has a descendant. A field typed as `Animal` accepts an `Animal` or descendant record. Validate and observe the assignable query:

```sh
mergeway-cli --root examples/inheritance validate
mergeway-cli --root examples/inheritance list --type Animal
mergeway-cli --root examples/inheritance get --type Animal dog-1
```

## Reference unrelated entity types

In `examples/reference-union`, `Activity.owner` accepts either a `User` or a `Team`:

```yaml
owner:
  type: User | Team
```

Verify both target kinds:

```sh
mergeway-cli --root examples/reference-union validate
mergeway-cli --root examples/reference-union list --type Activity
mergeway-cli --root examples/reference-union get --type Activity activity-2
```

The identifier must resolve in exactly one target set. Reusing the same owner identifier in both `User` and `Team` makes the union ambiguous and validation fails.

Inheritance is limited to native fields and one parent. Use [Identifiers and references](../../../reference/specification/identifiers-references/) for assignability and ambiguity rules, and [Entity definitions](../../../reference/specification/entities/) for inheritance constraints.
