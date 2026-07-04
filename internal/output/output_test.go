package output

import (
	"bytes"
	"strings"
	"testing"
)

type item struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

func (i item) Concise() string { return i.Name + " (" + i.ID + ")" }

var sample = []item{
	{Name: "Alice", ID: "444444444444444444"},
	{Name: "Bob", ID: "444444444444444445"},
}

func TestEmitConcise(t *testing.T) {
	var buf bytes.Buffer
	if err := Emit(&buf, "concise", sample); err != nil {
		t.Fatal(err)
	}
	want := "Alice (444444444444444444)\nBob (444444444444444445)\n"
	if buf.String() != want {
		t.Fatalf("concise = %q, want %q", buf.String(), want)
	}
}

func TestEmitJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := Emit(&buf, "json", sample); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"name": "Alice"`) {
		t.Fatalf("json = %s", buf.String())
	}
}

func TestEmitJSONL(t *testing.T) {
	var buf bytes.Buffer
	if err := Emit(&buf, "jsonl", sample); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 || !strings.HasPrefix(lines[0], `{"name":"Alice"`) {
		t.Fatalf("jsonl = %q", buf.String())
	}
}

func TestEmitTable(t *testing.T) {
	var buf bytes.Buffer
	if err := Emit(&buf, "table", sample); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "Alice") {
		t.Fatalf("table = %q", out)
	}
}

func TestEmitUnknownFormat(t *testing.T) {
	if err := Emit(&bytes.Buffer{}, "yaml", sample); err == nil {
		t.Fatal("unknown format must error")
	}
}

func TestEmitNonSlice(t *testing.T) {
	if err := Emit(&bytes.Buffer{}, "concise", "not-a-slice"); err == nil {
		t.Fatal("non-slice must error")
	}
}
