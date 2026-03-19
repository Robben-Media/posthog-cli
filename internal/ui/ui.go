package ui

import (
	"fmt"
	"os"

	"github.com/muesli/termenv"
)

var output = termenv.NewOutput(os.Stderr)

// Bold returns the text in bold.
func Bold(s string) string {
	return output.String(s).Bold().String()
}

// Dim returns the text dimmed.
func Dim(s string) string {
	return output.String(s).Faint().String()
}

// Green returns the text in green.
func Green(s string) string {
	return output.String(s).Foreground(output.Color("2")).String()
}

// Red returns the text in red.
func Red(s string) string {
	return output.String(s).Foreground(output.Color("1")).String()
}

// Yellow returns the text in yellow.
func Yellow(s string) string {
	return output.String(s).Foreground(output.Color("3")).String()
}

// Cyan returns the text in cyan.
func Cyan(s string) string {
	return output.String(s).Foreground(output.Color("6")).String()
}

// Info prints an informational message to stderr.
func Info(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// Warn prints a warning message to stderr.
func Warn(format string, args ...any) {
	fmt.Fprintf(os.Stderr, Yellow("warning: ")+format+"\n", args...)
}

// Error prints an error message to stderr.
func Error(format string, args ...any) {
	fmt.Fprintf(os.Stderr, Red("error: ")+format+"\n", args...)
}
