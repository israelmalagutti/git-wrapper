package stack

import (
	"fmt"
	"strings"

	"github.com/israelmalagutti/git-wrapper/internal/colors"
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

	depth := s.GetStackDepth(node.Name)
	chars := colors.DefaultTreeChars()

	// Prepare the branch line
	connector := chars.Tee + chars.Horizontal + chars.Horizontal
	if isLast {
		connector = chars.Corner + chars.Horizontal + chars.Horizontal
	}

	// Special handling for trunk (root)
	if node.Parent == nil {
		connector = chars.Bullet
		prefix = ""
	}

	// Color the connector based on depth
	coloredConnector := colors.CycleText(connector, depth)

	// Format branch name with appropriate color
	var branchName string
	if node.IsCurrent {
		branchName = colors.BranchCurrent(node.Name)
	} else if node.IsTrunk {
		branchName = colors.BranchTrunk(node.Name)
	} else {
		branchName = colors.CycleText(node.Name, depth)
	}

	// Build the line
	result.WriteString(prefix)
	result.WriteString(coloredConnector)
	result.WriteString(" ")
	result.WriteString(branchName)

	// Add trunk indicator
	if node.IsTrunk {
		result.WriteString(colors.Muted(" (trunk)"))
	}

	// Add commit SHA if requested
	if opts.ShowCommitSHA && node.CommitSHA != "" {
		result.WriteString(colors.Muted(fmt.Sprintf(" [%s]", node.CommitSHA[:7])))
	}

	// Add commit message if requested and detailed
	if opts.ShowCommitMsg && repo != nil {
		msg, err := repo.RunGitCommand("log", "-1", "--format=%s", node.Name)
		if err == nil && msg != "" {
			// Truncate long messages
			if len(msg) > 60 {
				msg = msg[:57] + "..."
			}
			result.WriteString(colors.DimText(fmt.Sprintf(" - %s", msg)))
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
				childPrefix = colors.CycleText(chars.Vertical, depth) + "   "
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

	chars := colors.DefaultTreeChars()
	indent := strings.Repeat("  ", depth)

	// Use filled circle for current branch, hollow for others
	indicator := chars.Circle
	if node.IsCurrent {
		indicator = chars.Bullet
	}

	// Color the indicator based on depth
	coloredIndicator := colors.CycleText(indicator, depth)

	// Format branch name with appropriate color
	var branchName string
	if node.IsCurrent {
		branchName = colors.BranchCurrent(node.Name)
	} else if node.IsTrunk {
		branchName = colors.BranchTrunk(node.Name)
	} else {
		branchName = colors.CycleText(node.Name, depth)
	}

	suffix := ""
	if node.IsTrunk {
		suffix = colors.Muted(" (trunk)")
	}

	result.WriteString(fmt.Sprintf("%s%s %s%s\n", indent, coloredIndicator, branchName, suffix))

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
			result.WriteString(colors.Muted(" → "))
		}

		// Color based on position
		var name string
		if node.IsCurrent {
			name = colors.BranchCurrent(node.Name)
		} else if node.IsTrunk {
			name = colors.BranchTrunk(node.Name)
		} else {
			name = colors.CycleText(node.Name, i)
		}
		result.WriteString(name)
	}

	return result.String()
}
