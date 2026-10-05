package diagram

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestRenderTextFlowchart(t *testing.T) {
	got, err := RenderText("graph TD\n  A[Write] --> B[Think]", 60)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(got.Lines, "\n")
	if got.Kind != "flowchart" || !strings.Contains(joined, "Write") || !strings.Contains(joined, "▼") {
		t.Fatalf("unexpected render (%s):\n%s", got.Kind, joined)
	}
}

func TestRenderTextSequence(t *testing.T) {
	got, err := RenderText("sequenceDiagram\n  TUI->>Worker: poll", 60)
	if err != nil || got.Kind != "sequence" || !strings.Contains(strings.Join(got.Lines, "\n"), "poll") {
		t.Fatalf("got %+v err=%v", got, err)
	}
}

func TestRenderTextUnsupported(t *testing.T) {
	got, err := RenderText("pie title Pets\n  \"Dogs\" : 3", 60)
	if err == nil {
		t.Fatalf("pie charts should not render as text")
	}
	if got.Kind != "pie" {
		t.Fatalf("kind = %q", got.Kind)
	}
}

func TestRenderTextClampsWidth(t *testing.T) {
	src := "graph LR\n  A[Alpha] --> B[Bravo] --> C[Charlie] --> D[Delta] --> E[Echo] --> F[Foxtrot]"
	got, err := RenderText(src, 30)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range got.Lines {
		if ansi.StringWidth(line) > 30 {
			t.Fatalf("line wider than limit: %q", line)
		}
	}
}
