package colors

import (
	"fmt"
	"os"
)

// Splog provides structured logging with semantic colors.
// Inspired by charcoal/graphite CLI's splog utility.
type Splog struct {
	quiet bool
	debug bool
}

// NewSplog creates a new Splog instance
func NewSplog() *Splog {
	return &Splog{
		quiet: false,
		debug: os.Getenv("GW_DEBUG") != "",
	}
}

// SetQuiet enables/disables quiet mode (suppresses info output)
func (s *Splog) SetQuiet(quiet bool) {
	s.quiet = quiet
}

// SetDebug enables/disables debug output
func (s *Splog) SetDebug(debug bool) {
	s.debug = debug
}

// Newline prints an empty line
func (s *Splog) Newline() {
	if s.quiet {
		return
	}
	fmt.Println()
}

// Print prints text without any formatting
func (s *Splog) Print(text string) {
	if s.quiet {
		return
	}
	fmt.Print(text)
}

// Println prints text with a newline
func (s *Splog) Println(text string) {
	if s.quiet {
		return
	}
	fmt.Println(text)
}

// Infof prints formatted informational text
func (s *Splog) Infof(format string, a ...interface{}) {
	if s.quiet {
		return
	}
	fmt.Printf(format, a...)
}

// Successf prints formatted success message (green)
func (s *Splog) Successf(format string, a ...interface{}) {
	if s.quiet {
		return
	}
	msg := fmt.Sprintf(format, a...)
	fmt.Print(Success(msg))
}

// Errorf prints formatted error message (red) to stderr
func (s *Splog) Errorf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprint(os.Stderr, Error("error: ")+msg)
}

// Warnf prints formatted warning message (yellow)
func (s *Splog) Warnf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Print(Warning("warning: ") + msg)
}

// Debugf prints formatted debug message (dim) if debug mode is enabled
func (s *Splog) Debugf(format string, a ...interface{}) {
	if !s.debug {
		return
	}
	msg := fmt.Sprintf(format, a...)
	fmt.Print(DimText("[debug] " + msg))
}

// Tipf prints formatted tip message (gray/muted)
func (s *Splog) Tipf(format string, a ...interface{}) {
	if s.quiet {
		return
	}
	msg := fmt.Sprintf(format, a...)
	fmt.Print(Muted("tip: " + msg))
}

// Plain prints plain text (respects quiet mode)
func (s *Splog) Plain(text string) {
	if s.quiet {
		return
	}
	fmt.Print(text)
}

// Package-level convenience functions using a default Splog instance

var defaultSplog = NewSplog()

// PrintSuccess prints a success message
func PrintSuccess(format string, a ...interface{}) {
	defaultSplog.Successf(format, a...)
}

// PrintError prints an error message
func PrintError(format string, a ...interface{}) {
	defaultSplog.Errorf(format, a...)
}

// PrintWarning prints a warning message
func PrintWarning(format string, a ...interface{}) {
	defaultSplog.Warnf(format, a...)
}

// PrintInfo prints an info message
func PrintInfo(format string, a ...interface{}) {
	defaultSplog.Infof(format, a...)
}

// PrintDebug prints a debug message
func PrintDebug(format string, a ...interface{}) {
	defaultSplog.Debugf(format, a...)
}

// Navigation feedback - inspired by charcoal's branch_traversal.ts

// PrintNav prints navigation feedback (arrow + branch name)
func PrintNav(direction, branchName string) {
	arrow := Success("→ ")
	switch direction {
	case "up":
		fmt.Println(arrow + BranchChild(branchName))
	case "down":
		fmt.Println(arrow + BranchParent(branchName))
	default:
		fmt.Println(arrow + Info(branchName))
	}
}

// PrintCheckout prints checkout feedback
func PrintCheckout(branchName string) {
	fmt.Println(Success("✓ ") + "Switched to " + BranchCurrent(branchName))
}

// PrintCreated prints branch creation feedback
func PrintCreated(branchName, parentName string) {
	fmt.Println(Success("✓ ") + "Created " + BranchCurrent(branchName) + " from " + BranchParent(parentName))
}

// PrintTracked prints branch tracking feedback
func PrintTracked(branchName, parentName string) {
	fmt.Println(Success("✓ ") + "Tracking " + BranchCurrent(branchName) + " with parent " + BranchParent(parentName))
}

// PrintDeleted prints branch deletion feedback
func PrintDeleted(branchName string) {
	fmt.Println(Success("✓ ") + "Deleted branch " + Muted(branchName))
}

// PrintRestacked prints restack feedback
func PrintRestacked(branchName, ontoName string) {
	fmt.Println(Success("✓ ") + "Restacked " + BranchCurrent(branchName) + " onto " + BranchParent(ontoName))
}

// PrintAlreadyUpToDate prints "already up to date" message
func PrintAlreadyUpToDate(branchName string) {
	fmt.Println(Info(branchName) + " is already up to date")
}

// PrintConflict prints conflict message
func PrintConflict(branchName, ontoName string) {
	fmt.Println(Warning("⚠ ") + "Conflict restacking " + Warning(branchName) + " onto " + Info(ontoName))
}
