---
title: "Using JSON Schema & JSONPath in Mergeway"
linkTitle: "JSONSchema & JSONPath"
description: "Advanced guide on using JSON Schema and JSON Path for setting up Mergeway with existing data structures"
weight: 30
---

Mergeway’s CLI lets you organize data from multiple sources into well‑typed entities.  While simple projects can define fields inline, larger teams often prefer to **re‑use existing JSON Schemas** or load many records from a single JSON document.  This guide shows how to install the CLI, author a JSON Schema‑backed entity, and leverage JSONPath selectors to extract multiple objects from a JSON file.

## Prerequisites

Make sure you have the latest Mergeway CLI installed before working with schemas. Review the [installation instructions](../getting-started/installation.md) for more details.

## Why JSON Schema?

A Mergeway entity typically lists fields manually, but this can become repetitive if you already maintain JSON Schemas elsewhere.  Mergeway supports a `json_schema` property that points at an on‑disk JSON Schema document.  The path is resolved relative to the file that declares the entity and must live inside the repository—external `$ref` documents are disallowed.  When `json_schema` is present you omit the `fields` map; Mergeway parses the schema and derives fields automatically.  Arrays become repeated fields, objects produce nested groups, and enumerations translate into enum types.  Keep schema files small and focus on one entity per file.

## Example: Customer Entity Backed by JSON Schema

1. **Define a JSON Schema.**  Create `schemas/customer.json` describing the shape of a customer record.  The sample below shows required fields (`id`, `name`, `tier`), enumerates allowed billing tiers (`trial`, `starter`, `enterprise`), and defines an array of `contacts` with email validation.

   ```jsonc
   {
     "$schema": "https://json-schema.org/draft/2020-12/schema",
     "type": "object",
     "description": "Customer metadata defined with JSON Schema",
     "required": ["id", "name", "tier"],
     "properties": {
       "id": { "type": "string" },
       "name": { "type": "string" },
       "tier": {
         "oneOf": [
           { "const": "trial" },
           { "const": "starter" },
           { "const": "enterprise" }
         ]
       },
       "contacts": {
         "type": "array",
         "items": {
           "type": "object",
           "required": ["label", "email"],
           "properties": {
             "label": { "type": "string" },
             "email": { "type": "string", "format": "email" }
           }
         }
       }
     }
   }
   ```

2. **Reference the schema in your workspace.**  Create a `mergeway.yaml` that declares a `Customer` entity and points at the schema file.  The `include` pattern tells Mergeway where to find customer records.

   ```yaml
   mergeway:
     version: 1

   entities:
     Customer:
       description: Customers defined via a JSON Schema file.
       identifier: id
       include:
         - data/customers/*.yaml
       json_schema: schemas/customer.json
   ```

3. **Add data files.**  Each YAML file under `data/customers/` should match the schema.  For example:

   ```yaml
   # data/customers/customer-001.yaml
   id: cust-001
   name: Example Industries
   tier: enterprise
   contacts:
     - label: Primary
       email: ops@example.com
     - label: Billing
       email: finance@example.com
   ```

4. **Validate and explore.**  Run `mergeway-cli validate` to ensure your data conforms to the schema.  Use `mergeway-cli data list --type Customer` to see loaded customers or `mergeway-cli gen-erd` to visualize relationships.

JSON Schema centralizes validation rules and makes your data contracts explicit.  Mergeway automatically converts nested objects into nested fields, sets array fields as repeated, and honours enum constraints.  This reduces duplication and lets you re‑use the same schema in other tools.

## Loading Multiple Records with JSONPath

Sometimes you want to load many objects from a single JSON file.  Instead of splitting the file, you can provide a JSONPath selector in your `include` mapping.  When several objects live in one file, use a selector to extract them.

## Users from a JSON Array

1. **Prepare the source JSON.**  Create `data/users.json` containing an array of users:

   ```json
   {
     "users": [
       { "id": "User-001", "name": "Ada" },
       { "id": "User-002", "name": "Grace" }
     ]
   }
   ```

2. **Define the entity with a JSONPath selector.**  In `mergeway.yaml`, specify the `path` and `selector` keys.  The `selector` uses JSONPath syntax (`$.users[*]`) to iterate over the elements of the `users` array.

   ```yaml
   mergeway:
     version: 1

   entities:
     User:
       identifier: id
       fields:
         id: string
         name: string
       include:
         - path: data/users.json
           selector: "$.users[*]"
   ```

3. **Load and validate.**  Run `mergeway-cli validate` to confirm the JSONPath extracts objects (non‑objects trigger a format error).  Then list users with `mergeway-cli data list User`.

**Tip:** Without a `selector`, Mergeway treats the entire JSON file as a single record or reads the `items` array if present. 
Selectors allow you to re-use existing data files / exports.

## Best Practices

* Store each entity’s schema in its own file and keep data organized in predictable folders (e.g., `data/customers/`, `data/users/`).
* Run `mergeway-cli validate` after every change to catch schema or data errors early.
* Use enums (`enum`/`const`/`oneOf`) and `format` constraints in your JSON Schema to enforce consistent values.
* Combine JSON Schema and JSONPath when dealing with complex documents—derive field definitions from the schema and extract many objects with selectors.

By integrating JSON Schema and JSONPath into your Mergeway workspace you can model rich data structures, enforce contracts, and scale your repository organization.

