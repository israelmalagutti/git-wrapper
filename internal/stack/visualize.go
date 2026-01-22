package stack

import (
	"fmt"
	"sort"
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

// Commit represents a single commit in a branch
type Commit struct {
	SHA     string
	Message string
}

// RenderTree renders the stack as a top-down tree with commits
// Output flows from leaves (top) down to trunk (bottom)
func (s *Stack) RenderTree(repo *git.Repo, opts TreeOptions) string {
	var result strings.Builder

	// Get ordered branches from leaves to trunk
	orderedBranches := s.getTopDownOrder()

	for i, node := range orderedBranches {
		isLast := i == len(orderedBranches)-1
		s.renderBranchWithCommits(&result, node, repo, opts, isLast)
	}

	return result.String()
}

// getTopDownOrder returns branches ordered from leaves to trunk
// Handles multiple stacks by rendering them in sequence
func (s *Stack) getTopDownOrder() []*Node {
	// Find all leaves (branches with no children)
	var leaves []*Node
	for _, node := range s.Nodes {
		if len(node.Children) == 0 && !node.IsTrunk {
			leaves = append(leaves, node)
		}
	}

	// Sort leaves alphabetically for consistent ordering
	sort.Slice(leaves, func(i, j int) bool {
		return leaves[i].Name < leaves[j].Name
	})

	// If no leaves, trunk is the only branch
	if len(leaves) == 0 {
		return []*Node{s.Trunk}
	}

	// Build ordered list by traversing from each leaf to trunk
	// But DON'T add trunk yet - we'll add it at the very end
	var ordered []*Node
	seen := make(map[string]bool)
	seen[s.Trunk.Name] = true // Mark trunk as seen so we skip it during traversal

	for _, leaf := range leaves {
		// Get path from leaf to trunk
		path := s.getPathToTrunk(leaf)

		// Add branches we haven't seen yet (except trunk)
		for _, node := range path {
			if !seen[node.Name] {
				seen[node.Name] = true
				ordered = append(ordered, node)
			}
		}
	}

	// Add trunk at the very end
	ordered = append(ordered, s.Trunk)

	return ordered
}

// getPathToTrunk returns path from a node to trunk (leaf first, trunk last)
func (s *Stack) getPathToTrunk(node *Node) []*Node {
	var path []*Node
	current := node
	for current != nil {
		path = append(path, current)
		current = current.Parent
	}
	return path
}

// renderBranchWithCommits renders a branch and its commits
func (s *Stack) renderBranchWithCommits(result *strings.Builder, node *Node, repo *git.Repo, opts TreeOptions, isLast bool) {
	chars := colors.DefaultTreeChars()
	depth := s.GetStackDepth(node.Name)

	// Use filled circle for current branch, hollow for others
	indicator := chars.Circle // ○
	if node.IsCurrent {
		indicator = chars.FilledCircle // ◉
	}

	// Format indicator and branch name
	// Current branch: bright/saturated, others: muted gray
	var coloredIndicator, branchName string
	if node.IsCurrent {
		coloredIndicator = colors.CycleText(indicator, depth)
		branchName = colors.BranchCurrent(node.Name)
	} else {
		coloredIndicator = colors.Muted(indicator)
		branchName = colors.Muted(node.Name)
	}

	// Build the branch line
	result.WriteString(coloredIndicator)
	result.WriteString(" ")
	result.WriteString(branchName)

	// Add current indicator
	if node.IsCurrent {
		result.WriteString(colors.Muted(" (current)"))
	}

	// Get commits for this branch
	var commits []Commit
	if repo != nil {
		if node.IsTrunk {
			// For trunk, show recent commits (last 3)
			commits = getTrunkCommits(repo, node.Name, 3)
		} else {
			// For other branches, show commits unique to this branch
			commits = s.getBranchCommits(repo, node)
		}

		// Add time since last commit (only if there are commits)
		if len(commits) > 0 {
			timeAgo := getTimeSinceLastCommit(repo, node.Name)
			if timeAgo != "" {
				result.WriteString(colors.Muted(" · " + timeAgo))
			}
		}
	}

	result.WriteString("\n")

	// Render commits
	if repo != nil {

		// White vertical line for commits
		verticalLine := colors.Muted(chars.Vertical)

		for i, commit := range commits {
			result.WriteString(verticalLine)
			result.WriteString(" ")

			// Commit SHA (shortened)
			sha := commit.SHA
			if len(sha) > 7 {
				sha = sha[:7]
			}

			// Only color the first (latest) commit SHA
			if i == 0 {
				result.WriteString(colors.CommitSHA(sha))
			} else {
				result.WriteString(colors.Muted(sha))
			}
			result.WriteString(colors.Muted(" - "))

			// Commit message (truncated)
			msg := commit.Message
			if len(msg) > 50 {
				msg = msg[:47] + "..."
			}
			result.WriteString(colors.Muted(msg))
			result.WriteString("\n")
		}

		// Add trailing vertical line (white) if there are commits or if not trunk
		if len(commits) > 0 || !node.IsTrunk {
			result.WriteString(verticalLine)
			result.WriteString("\n")
		}
	} else if !node.IsTrunk {
		// No repo but not trunk - still show connector
		verticalLine := colors.Muted(chars.Vertical)
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
func (s *Stack) RenderShort() string {
	var result strings.Builder

	// Get ordered branches from leaves to trunk
	orderedBranches := s.getTopDownOrder()

	for _, node := range orderedBranches {
		s.renderShortBranch(&result, node)
	}

	return result.String()
}

// renderShortBranch renders a single branch in short format
func (s *Stack) renderShortBranch(result *strings.Builder, node *Node) {
	chars := colors.DefaultTreeChars()
	depth := s.GetStackDepth(node.Name)

	// Use filled circle for current branch, hollow for others
	indicator := chars.Circle
	if node.IsCurrent {
		indicator = chars.FilledCircle
	}

	// Format indicator and branch name
	// Current branch: bright/saturated, others: muted gray
	var coloredIndicator, branchName string
	if node.IsCurrent {
		coloredIndicator = colors.CycleText(indicator, depth)
		branchName = colors.BranchCurrent(node.Name)
	} else {
		coloredIndicator = colors.Muted(indicator)
		branchName = colors.Muted(node.Name)
	}

	// Build suffix with current indicator
	suffix := ""
	if node.IsCurrent {
		suffix += colors.Muted(" (current)")
	}

	result.WriteString(fmt.Sprintf("%s %s%s\n", coloredIndicator, branchName, suffix))

	// Add white vertical connector if not trunk
	if !node.IsTrunk {
		verticalLine := colors.Muted(chars.Vertical)
		result.WriteString(verticalLine + "\n")
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
