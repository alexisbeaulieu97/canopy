// Package output provides helpers for CLI output formatting.
package output

import "fmt"

// Success prints a success message in the format: "<action> <target>\n"
// Example: Success("Created workspace", "my-workspace") -> "Created workspace my-workspace"
func Success(action, target string) {
	fmt.Printf("%s %s\n", action, target) //nolint:forbidigo // user-facing CLI output
}

// SuccessWithPath prints a success message with a path in the format: "<action> <target> in <path>\n"
// Example: SuccessWithPath("Created workspace", "my-ws", "/path/to/ws") -> "Created workspace my-ws in /path/to/ws"
func SuccessWithPath(action, target, path string) {
	fmt.Printf("%s %s in %s\n", action, target, path) //nolint:forbidigo // user-facing CLI output
}

// Info prints a neutral information message.
// Example: Info("No orphaned worktrees found.") -> "No orphaned worktrees found."
func Info(message string) {
	fmt.Println(message) //nolint:forbidigo // user-facing CLI output
}

// Infof prints a formatted neutral information message.
// Example: Infof("Found %d items", 5) -> "Found 5 items"
func Infof(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...) //nolint:forbidigo // user-facing CLI output
}

// Warn prints a warning message.
// Example: Warn("Configuration may be incomplete") -> "Configuration may be incomplete"
func Warn(message string) {
	fmt.Println(message) //nolint:forbidigo // user-facing CLI output
}

// Warnf prints a formatted warning message.
// Example: Warnf("Missing %d files", 3) -> "Missing 3 files"
func Warnf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...) //nolint:forbidigo // user-facing CLI output
}

// Print prints a message without newline.
// Use for raw output like paths or data.
func Print(message string) {
	fmt.Print(message) //nolint:forbidigo // user-facing CLI output
}

// Printf prints a formatted message without newline.
func Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...) //nolint:forbidigo // user-facing CLI output
}

// Println prints a message with newline.
func Println(message string) {
	fmt.Println(message) //nolint:forbidigo // user-facing CLI output
}
