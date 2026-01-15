package stack

import (
	"fmt"
	"strings"

	"github.com/israelmalagutti/git-wrapper/internal/git"
)

// TreeOptions controls how the tree is rendered
type TreeOptions struct {
	ShowCommitSHA bool
	ShowCommitMsg bool
	Detailed      bool
}

// RenderTree renders the stack as an ASCII tree
func (s *Stack) RenderTree(repo *git.Repo, opts TreeOptions) string {
	var result strings.Builder

	// Render from trunk down
	s.renderNode(&result, s.Trunk, "", true, repo, opts)

	return result.String()
}

func (s *Stack) renderNode(result *strings.Builder, node *Node, prefix string, isLast bool, repo *git.Repo, opts TreeOptions) {
	if node == nil {
		return
	}

	// Prepare the branch line
	connector := "├──"
	if isLast {
		connector = "└──"
	}

	// Special handling for trunk (root)
	if node.Parent == nil {
		connector = "●"
		prefix = ""
	}

	// Branch indicator (appears right before branch name)
	indicator := ""
	if node.IsCurrent {
		indicator = "*"
	}

	// Build the line
	line := fmt.Sprintf("%s%s %s%s", prefix, connector, indicator, node.Name)

	// Add trunk indicator
	if node.IsTrunk {
		line += " (trunk)"
	}

	result.WriteString(line)

	// Add commit SHA if requested
	if opts.ShowCommitSHA && node.CommitSHA != "" {
		result.WriteString(fmt.Sprintf(" [%s]", node.CommitSHA[:7]))
	}

	// Add commit message if requested and detailed
	if opts.ShowCommitMsg && repo != nil {
		msg, err := repo.RunGitCommand("log", "-1", "--format=%s", node.Name)
		if err == nil && msg != "" {
			// Truncate long messages
			if len(msg) > 60 {
				msg = msg[:57] + "..."
			}
			result.WriteString(fmt.Sprintf(" - %s", msg))
		}
	}

	result.WriteString("\n")

	// Render children
	childCount := len(node.Children)
	for i, child := range node.Children {
		isLastChild := i == childCount-1

		// Prepare prefix for children
		var childPrefix string
		if node.Parent == nil {
			// Trunk level
			childPrefix = ""
		} else {
			if isLast {
				childPrefix = prefix + "    "
			} else {
				childPrefix = prefix + "│   "
			}
		}

		s.renderNode(result, child, childPrefix, isLastChild, repo, opts)
	}
}

// RenderShort renders a compact view of the stack
func (s *Stack) RenderShort() string {
	var result strings.Builder

	// Simple list view
	s.renderShortNode(&result, s.Trunk, 0)

	return result.String()
}

func (s *Stack) renderShortNode(result *strings.Builder, node *Node, depth int) {
	if node == nil {
		return
	}

	indent := strings.Repeat("  ", depth)

	// Use filled circle for current branch, hollow for others
	indicator := "○"
	if node.IsCurrent {
		indicator = "●"
	}

	suffix := ""
	if node.IsTrunk {
		suffix = " (trunk)"
	}

	result.WriteString(fmt.Sprintf("%s%s %s%s\n", indent, indicator, node.Name, suffix))

	// Render children
	for _, child := range node.Children {
		s.renderShortNode(result, child, depth+1)
	}
}

// RenderPath renders a path from trunk to a branch
func (s *Stack) RenderPath(branch string) string {
	path := s.FindPath(branch)
	if path == nil {
		return ""
	}

	var result strings.Builder
	for i, node := range path {
		if i > 0 {
			result.WriteString(" → ")
		}
		result.WriteString(node.Name)
	}

	return result.String()
}
