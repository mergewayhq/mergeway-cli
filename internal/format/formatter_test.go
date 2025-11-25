package format

import (
	"strings"
	"testing"
)

func TestFormatBytesSortsItems(t *testing.T) {
	input := `
type: Post
items:
  - id: post-b
    title: Beta
  - id: post-a
    title: Alpha
`

	out, err := FormatBytes("data.yaml", []byte(input), nil)
	if err != nil {
		t.Fatalf("format bytes: %v", err)
	}

	expected := `type: Post
items:
  - id: post-a
    title: Alpha
  - id: post-b
    title: Beta
`

	if got := string(out); strings.TrimSpace(got) != strings.TrimSpace(expected) {
		t.Fatalf("unexpected output:\n%s", got)
	}
}

func TestFormatBytesPreservesComments(t *testing.T) {
	input := `
items:
  # first block
  - id: item-b
    title: Bravo
  # important comment
  - id: item-a
    title: Alpha
`

	out, err := FormatBytes("data.yaml", []byte(input), nil)
	if err != nil {
		t.Fatalf("format bytes: %v", err)
	}

	body := string(out)
	if !strings.Contains(body, "# important comment\n  - id: item-a") {
		t.Fatalf("expected comment to follow item-a, got:\n%s", body)
	}
	if !strings.Contains(body, "# first block\n  - id: item-b") {
		t.Fatalf("expected comment to follow item-b, got:\n%s", body)
	}
}

func TestFormatBytesSupportsJSON(t *testing.T) {
	input := `{
  "type": "Post",
  "items": [
    {"id": "post-b", "title": "Beta"},
    {"id": "post-a", "title": "Alpha"}
  ]
}`

	out, err := FormatBytes("data.json", []byte(input), nil)
	if err != nil {
		t.Fatalf("format bytes json: %v", err)
	}

	body := string(out)
	if !strings.Contains(body, `"id": "post-a"`) || !strings.Contains(body, `"id": "post-b"`) {
		t.Fatalf("unexpected json output:\n%s", body)
	}
	first := strings.Index(body, `"id": "post-a"`)
	second := strings.Index(body, `"id": "post-b"`)
	if first < 0 || second < 0 || first > second {
		t.Fatalf("expected post-a before post-b, got:\n%s", body)
	}
}

func TestFormatBytesOrdersFieldsUsingSchema(t *testing.T) {
	input := `type: Post
items:
  - title: Title
    id: post-2
    meta:
      summary: B
      slug: b
    author: user-1
`
	schema := NewSchema([]*SchemaField{
		{Name: "id"},
		{Name: "title"},
		{
			Name: "meta",
			Nested: NewSchema([]*SchemaField{
				{Name: "slug"},
				{Name: "summary"},
			}),
		},
		{Name: "author"},
	})

	out, err := FormatBytes("data.yaml", []byte(input), schema)
	if err != nil {
		t.Fatalf("format bytes with schema: %v", err)
	}
	body := string(out)
	if idxID := strings.Index(body, "id:"); idxID == -1 {
		t.Fatalf("expected id field in output:\n%s", body)
	}
	idxID := strings.Index(body, "id: post-2")
	idxTitle := strings.Index(body, "title: Title")
	idxAuthor := strings.Index(body, "author: user-1")
	if idxID >= idxTitle || idxTitle >= idxAuthor {
		t.Fatalf("expected order id -> title -> author, got:\n%s", body)
	}
	if !strings.Contains(body, "meta:\n      slug: b\n      summary: B") {
		t.Fatalf("expected nested meta fields ordered, got:\n%s", body)
	}
}
