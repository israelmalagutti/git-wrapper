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

// SiblingInfo tracks the position of a branch among its siblings
type SiblingInfo struct {
	IsFirst       bool // First sibling (topmost in output)
	IsLast        bool // Last sibling (closest to parent)
	IsOnlyChild   bool // Only child of parent (no junctions needed)
	HasMoreAbove  bool // Are there siblings above this one?
	HasMoreBelow  bool // Are there siblings below this one?
}

// Commit represents a single commit in a branch
type Commit struct {
	SHA     string
	Message string
}

// RenderTree renders the stack as a top-down tree with commits
// Output flows from leaves (top) down to trunk (bottom)
func (s *Stack) RenderTree(repo *git.Repo, opts TreeOptions) string {
	var result strings.Builder

	// Render trunk's children recursively, then trunk
	children := s.Trunk.SortedChildren()
	if len(children) > 0 {
		// Start with empty rail - the rail character is added inside
		s.renderSiblingsWithCommits(&result, children, repo, opts, "")
	}

	// Render trunk at the bottom
	s.renderTrunkWithCommits(&result, s.Trunk, repo, opts)

	return result.String()
}

// renderSiblingsWithCommits renders a group of siblings with a shared rail connecting to parent
// outerRail is the prefix for this sibling group (e.g., "│   " for test's children)
func (s *Stack) renderSiblingsWithCommits(result *strings.Builder, siblings []*Node, repo *git.Repo, opts TreeOptions, outerRail string) {
	chars := colors.DefaultTreeChars()
	vertLine := colors.Muted(chars.Vertical)

	for i, node := range siblings {
		isFirst := i == 0
		hasChildren := len(node.Children) > 0

		// Calculate the prefix for this node and its children
		// If node has children, they form a visual group with the same indentation
		var nodeRail string
		if hasChildren {
			// Node with children: indent the whole subtree (including the node itself)
			nodeRail = outerRail + vertLine + "   "
		} else {
			nodeRail = outerRail
		}

		// Render this node's children first (they appear above)
		if hasChildren {
			grandchildren := node.SortedChildren()
			s.renderSiblingsWithCommits(result, grandchildren, repo, opts, nodeRail)
		}

		// Build prefix for this node's branch line
		var branchPrefix string
		if isFirst && outerRail == "" && !hasChildren {
			// First sibling at root level that's a leaf - no prefix
			branchPrefix = ""
		} else if hasChildren {
			// Node with children uses the indented rail
			branchPrefix = nodeRail
		} else {
			// Leaf node uses outer rail
			branchPrefix = outerRail
		}

		// commitRail is what comes before the vertical line for commits/connectors
		var commitRail string
		if hasChildren {
			commitRail = nodeRail
		} else {
			commitRail = outerRail
		}

		s.renderBranchLineWithCommits(result, node, repo, opts, branchPrefix, commitRail)

		// Show merge junction if this node has children
		if hasChildren {
			result.WriteString(outerRail)
			result.WriteString(colors.Muted(chars.Tee + chars.Horizontal + chars.Horizontal + chars.Horizontal + chars.BottomRight))
			result.WriteString("\n")
		}
	}
}

// renderBranchLineWithCommits renders a single branch line and its commits
// branchPrefix is the prefix for the branch indicator line
// rail is the prefix for commit lines and connectors
func (s *Stack) renderBranchLineWithCommits(result *strings.Builder, node *Node, repo *git.Repo, opts TreeOptions, branchPrefix string, rail string) {
	chars := colors.DefaultTreeChars()
	depth := s.GetStackDepth(node.Name)

	// Use filled circle for current branch, hollow for others
	indicator := chars.Circle
	if node.IsCurrent {
		indicator = chars.FilledCircle
	}

	// Format indicator and branch name
	var coloredIndicator, branchName string
	if node.IsCurrent {
		coloredIndicator = colors.CycleText(indicator, depth)
		branchName = colors.BranchCurrent(node.Name)
	} else {
		coloredIndicator = colors.Muted(indicator)
		branchName = colors.Muted(node.Name)
	}

	// Build the branch line
	result.WriteString(branchPrefix)
	result.WriteString(coloredIndicator)
	result.WriteString(" ")
	result.WriteString(branchName)

	if node.IsCurrent {
		result.WriteString(colors.Muted(" (current)"))
	}

	// Get commits
	var commits []Commit
	if repo != nil {
		commits = s.getBranchCommits(repo, node)
		if len(commits) > 0 {
			timeAgo := getTimeSinceLastCommit(repo, node.Name)
			if timeAgo != "" {
				result.WriteString(colors.Muted(" · " + timeAgo))
			}
		}
	}

	result.WriteString("\n")

	// Render commits with rail prefix
	if repo != nil {
		for i, commit := range commits {
			result.WriteString(rail)
			result.WriteString(colors.Muted(chars.Vertical))
			result.WriteString(" ")

			sha := commit.SHA
			if len(sha) > 7 {
				sha = sha[:7]
			}
			if i == 0 {
				result.WriteString(colors.CommitSHA(sha))
			} else {
				result.WriteString(colors.Muted(sha))
			}
			result.WriteString(colors.Muted(" - "))

			msg := commit.Message
			if len(msg) > 50 {
				msg = msg[:47] + "..."
			}
			result.WriteString(colors.Muted(msg))
			result.WriteString("\n")
		}
	}

	// Trailing connector line
	result.WriteString(rail)
	result.WriteString(colors.Muted(chars.Vertical))
	result.WriteString("\n")
}

// renderTrunkWithCommits renders the trunk branch
func (s *Stack) renderTrunkWithCommits(result *strings.Builder, node *Node, repo *git.Repo, opts TreeOptions) {
	chars := colors.DefaultTreeChars()
	depth := s.GetStackDepth(node.Name)

	indicator := chars.Circle
	if node.IsCurrent {
		indicator = chars.FilledCircle
	}

	var coloredIndicator, branchName string
	if node.IsCurrent {
		coloredIndicator = colors.CycleText(indicator, depth)
		branchName = colors.BranchCurrent(node.Name)
	} else {
		coloredIndicator = colors.Muted(indicator)
		branchName = colors.Muted(node.Name)
	}

	result.WriteString(coloredIndicator)
	result.WriteString(" ")
	result.WriteString(branchName)

	if node.IsCurrent {
		result.WriteString(colors.Muted(" (current)"))
	}

	var commits []Commit
	if repo != nil {
		commits = getTrunkCommits(repo, node.Name, 3)
		if len(commits) > 0 {
			timeAgo := getTimeSinceLastCommit(repo, node.Name)
			if timeAgo != "" {
				result.WriteString(colors.Muted(" · " + timeAgo))
			}
		}
	}

	result.WriteString("\n")

	// Render trunk commits
	if repo != nil {
		verticalLine := colors.Muted(chars.Vertical)
		for i, commit := range commits {
			result.WriteString(verticalLine)
			result.WriteString(" ")

			sha := commit.SHA
			if len(sha) > 7 {
				sha = sha[:7]
			}
			if i == 0 {
				result.WriteString(colors.CommitSHA(sha))
			} else {
				result.WriteString(colors.Muted(sha))
			}
			result.WriteString(colors.Muted(" - "))

			msg := commit.Message
			if len(msg) > 50 {
				msg = msg[:47] + "..."
			}
			result.WriteString(colors.Muted(msg))
			result.WriteString("\n")
		}

		// Trailing connector for trunk
		result.WriteString(verticalLine)
		result.WriteString("\n")
	}
}

// getTimeSinceLastCommit returns relative time since the last commit on a branch
func getTimeSinceLastCommit(repo *git.Repo, branch string) string {
	output, err := repo.RunGitCommand("log", "-1", "--format=%cr", branch)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(output)
}

// getTrunkCommits returns the last n commits on trunk
func getTrunkCommits(repo *git.Repo, branch string, n int) []Commit {
	output, err := repo.RunGitCommand("log", "--oneline", fmt.Sprintf("-%d", n), branch)
	if err != nil || output == "" {
		return nil
	}

	var commits []Commit
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) >= 2 {
			commits = append(commits, Commit{
				SHA:     parts[0],
				Message: parts[1],
			})
		} else if len(parts) == 1 {
			commits = append(commits, Commit{
				SHA:     parts[0],
				Message: "",
			})
		}
	}

	return commits
}

// getBranchCommits returns the commits unique to this branch (not in parent)
func (s *Stack) getBranchCommits(repo *git.Repo, node *Node) []Commit {
	if node.Parent == nil {
		return nil
	}

	// Get commits in this branch that are not in parent
	// git log parent..branch --oneline
	output, err := repo.RunGitCommand("log", "--oneline", "--reverse",
		fmt.Sprintf("%s..%s", node.Parent.Name, node.Name))
	if err != nil || output == "" {
		return nil
	}

	var commits []Commit
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) >= 2 {
			commits = append(commits, Commit{
				SHA:     parts[0],
				Message: parts[1],
			})
		} else if len(parts) == 1 {
			commits = append(commits, Commit{
				SHA:     parts[0],
				Message: "",
			})
		}
	}

	return commits
}

// RenderShort renders a compact view of the stack (top-down, no commits)
// Uses T-junctions to show sibling relationships
// RenderShort renders a compact view of the stack (no commits)
func (s *Stack) RenderShort() string {
	var result strings.Builder
	chars := colors.DefaultTreeChars()

	// Render trunk's children recursively, then trunk
	children := s.Trunk.SortedChildren()
	if len(children) > 0 {
		s.renderSiblingsShort(&result, children, "")
	}

	// Render trunk at the bottom
	indicator := chars.Circle
	if s.Trunk.IsCurrent {
		indicator = chars.FilledCircle
	}
	var coloredIndicator, branchName string
	if s.Trunk.IsCurrent {
		coloredIndicator = colors.CycleText(indicator, 0)
		branchName = colors.BranchCurrent(s.Trunk.Name)
	} else {
		coloredIndicator = colors.Muted(indicator)
		branchName = colors.Muted(s.Trunk.Name)
	}
	result.WriteString(coloredIndicator)
	result.WriteString(" ")
	result.WriteString(branchName)
	if s.Trunk.IsCurrent {
		result.WriteString(colors.Muted(" (current)"))
	}
	result.WriteString("\n")

	return result.String()
}

// renderSiblingsShort renders siblings with a shared rail (short format)
func (s *Stack) renderSiblingsShort(result *strings.Builder, siblings []*Node, outerRail string) {
	chars := colors.DefaultTreeChars()
	vertLine := colors.Muted(chars.Vertical)

	for i, node := range siblings {
		isFirst := i == 0
		hasChildren := len(node.Children) > 0

		// Calculate the prefix for this node and its children
		var nodeRail string
		if hasChildren {
			nodeRail = outerRail + vertLine + "   "
		} else {
			nodeRail = outerRail
		}

		// Render children first (they appear above)
		if hasChildren {
			grandchildren := node.SortedChildren()
			s.renderSiblingsShort(result, grandchildren, nodeRail)
		}

		// Build prefix for branch line
		var branchPrefix string
		if isFirst && outerRail == "" && !hasChildren {
			// First sibling at root level that's a leaf - no prefix
			branchPrefix = ""
		} else if hasChildren {
			branchPrefix = nodeRail
		} else {
			branchPrefix = outerRail
		}

		// Render branch line
		depth := s.GetStackDepth(node.Name)
		indicator := chars.Circle
		if node.IsCurrent {
			indicator = chars.FilledCircle
		}
		var coloredIndicator, branchName string
		if node.IsCurrent {
			coloredIndicator = colors.CycleText(indicator, depth)
			branchName = colors.BranchCurrent(node.Name)
		} else {
			coloredIndicator = colors.Muted(indicator)
			branchName = colors.Muted(node.Name)
		}

		result.WriteString(branchPrefix)
		result.WriteString(coloredIndicator)
		result.WriteString(" ")
		result.WriteString(branchName)
		if node.IsCurrent {
			result.WriteString(colors.Muted(" (current)"))
		}
		result.WriteString("\n")

		// Connector line
		var connectorRail string
		if hasChildren {
			connectorRail = nodeRail
		} else {
			connectorRail = outerRail
		}
		result.WriteString(connectorRail)
		result.WriteString(vertLine)
		result.WriteString("\n")

		// Show merge junction if this node has children
		if hasChildren {
			result.WriteString(outerRail)
			result.WriteString(colors.Muted(chars.Tee + chars.Horizontal + chars.Horizontal + chars.Horizontal + chars.BottomRight))
			result.WriteString("\n")
		}
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
