package diagram

import (
	"fmt"
	"strings"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
	"github.com/AlexanderGrooff/mermaid-ascii/pkg/render"
	"github.com/charmbracelet/x/ansi"
)

type Text struct {
	Kind      string
	Lines     []string
	Truncated bool
}

func Kind(source string) string {
	for _, line := range strings.Split(source, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || strings.HasPrefix(fields[0], "%%") || fields[0] == "---" {
			continue
		}
		switch fields[0] {
		case "graph", "flowchart":
			return "flowchart"
		case "sequenceDiagram":
			return "sequence"
		case "erDiagram":
			return "entity relationship"
		}
		return strings.TrimSuffix(fields[0], "Diagram")
	}
	return "diagram"
}

func RenderText(source string, maxWidth int) (result Text, err error) {
	result.Kind = Kind(source)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("could not draw this %s", result.Kind)
		}
	}()
	cfg := diagram.DefaultConfig()
	cfg.MaxWidth = maxWidth
	out, renderErr := render.RenderDiagram(source, cfg)
	if renderErr != nil {
		return result, renderErr
	}
	for _, line := range strings.Split(strings.TrimRight(out, "\n "), "\n") {
		line = strings.TrimRight(line, " ")
		if ansi.StringWidth(line) > maxWidth {
			line = ansi.Truncate(line, maxWidth-1, "…")
			result.Truncated = true
		}
		result.Lines = append(result.Lines, line)
	}
	return result, nil
}
