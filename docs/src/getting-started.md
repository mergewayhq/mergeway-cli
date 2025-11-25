# Getting Started with Mergeway

Goal: scaffold a workspace, define an entity, evolve the layout as requirements grow, and learn the core Mergeway commands end-to-end.

> All commands assume the `mw` binary is on your `PATH`.

## 1. Scaffold a Workspace with Inline Data

```bash
mkdir farmers-market
cd farmers-market
mw init
```

`mw init` creates a `mergeway.yaml` entry file. Replace its contents with an inline entity that also carries a few inline records:

```yaml
mergeway:
  version: 1

entities:
  Category:
    description: Simple lookup table for product groupings
    identifier: slug
    fields:
      slug:
        type: string
        required: true
      label:
        type: string
        required: true
    data:
      - slug: produce
        label: Fresh Produce
      - slug: pantry
        label: Pantry Staples
```

Try a few commands:

```bash
mw entity list
mw entity show Category
mw list --type Category
mw validate
```

At this stage everything lives in a single file—perfect for tiny datasets.

## 2. Move Records into External YAML Files

As the table grows, shift the data into dedicated files. Create a folder for category data and move the records there:

```bash
mkdir -p data/categories
cat <<'YAML' > data/categories/categories.yaml
items:
  - slug: produce
    label: Fresh Produce
  - slug: pantry
    label: Pantry Staples
  - slug: beverages
    label: Beverages
YAML
```

Update `mergeway.yaml` so `Category` reads from the new file:

```yaml
mergeway:
  version: 1

entities:
  Category:
    description: Simple lookup table for product groupings
    identifier: slug
    include:
      - data/categories/*.yaml
    fields:
      slug:
        type: string
        required: true
      label:
        type: string
        required: true
```

Re-run the commands to see the effect:

```bash
mw list --type Category
mw get --type Category beverages
mw validate
```

## 3. Split Schema Definitions and Add JSON Data

Larger workspaces benefit from keeping schemas in their own files. Create an `entities/` folder for additional definitions:

```bash
mkdir -p entities
```

Add a new `Product` entity that pulls from a JSON file using a JSONPath selector:

```bash
cat <<'YAML' > entities/Product.yaml
mergeway:
  version: 1

entities:
  Product:
    description: Market products with category references
    identifier: sku
    include:
      - path: data/products.json
        selector: "$.items[*]"
    fields:
      sku:
        type: string
        required: true
      name:
        type: string
        required: true
      category:
        type: Category
        required: true
      price:
        type: number
        required: true
YAML
```

Create the JSON data file the schema expects. Notice that one product references a `household` category that we haven't defined yet:

```bash
cat <<'JSON' > data/products.json
{
  "items": [
    {"sku": "apple-001", "name": "Honeycrisp Apple", "category": "produce", "price": 1.25},
    {"sku": "oat-500", "name": "Rolled Oats", "category": "pantry", "price": 4.99},
    {"sku": "soap-010", "name": "Castile Soap", "category": "household", "price": 6.75}
  ]
}
JSON
```

Finally, have `mergeway.yaml` pull in any external schemas:

```yaml
mergeway:
  version: 1

include:
  - entities/*.yaml

entities:
  Category:
    description: Simple lookup table for product groupings
    identifier: slug
    include:
      - data/categories/*.yaml
    fields:
      slug:
        type: string
        required: true
      label:
        type: string
        required: true
```

Explore the richer workspace:

```bash
mw entity list
mw entity show Product
mw list --type Product
mw validate
```

`mw validate` now reports a broken reference because the `household` category doesn't exist yet:

```
phase: references
type: Product
id: soap-010
file: data/products.json
message: referenced Category "household" not found
```

Add the missing category to the YAML file and validate again:

```bash
cat <<'YAML' >> data/categories/categories.yaml
- slug: household
  label: Household Goods
YAML

mw validate
```

With the additional category in place, validation succeeds and both entities are in sync.

## 4. Export Everything as JSON

Collect the full dataset into a single snapshot:

```bash
mw export --format json --output market-snapshot.json
cat market-snapshot.json
```

## Keep the Workflow Running Smoothly

Once the basics feel comfortable, automate formatting and reviews so the workspace stays healthy:

- [Set Up Mergeway with GitHub](how-to/setup-mergeway-github.md) to enforce `mw fmt --lint` in Actions and route reviews through CODEOWNERS.
- [Enforce Mergeway Formatting with pre-commit](how-to/setup-mergeway-pre-commit.md) so contributors run `mw fmt` locally before every commit.

## You're Done!

Nice work—you’ve defined entities inline, moved data to YAML, added JSON-backed entities, and exercised the key Mergeway commands. You're ready to scale the workspace to your team’s needs.
