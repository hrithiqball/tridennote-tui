package tui

import (
	"os"

	"github.com/atotto/clipboard"
	osc52 "github.com/aymanbagabas/go-osc52/v2"
)

var copyToClipboard = func(text string) error {
	if err := clipboard.WriteAll(text); err == nil {
		return nil
	}
	_, err := osc52.New(text).WriteTo(os.Stderr)
	return err
}
