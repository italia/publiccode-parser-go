package git

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// Provides functionality for cloning and checking files in Git repositories.
type GitHelper struct {
	// Base directory for all temporary Git clones
	tempDir string
	// Maps repository URLs to their local clone paths
	ClonedRepos map[string]string
}

func NewGitHelper() (*GitHelper, error) {
	tempDir, err := os.MkdirTemp("", "publiccode-git-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return &GitHelper{
		tempDir:     tempDir,
		ClonedRepos: make(map[string]string),
	}, nil
}

// Performs a sparse clone of a Git repository.
func (g *GitHelper) CloneRepo(repoURL string) (string, error) {
	// Check if already cloned
	if clonePath, ok := g.ClonedRepos[repoURL]; ok {
		return clonePath, nil
	}

	// Create a repo name
	repoName := strings.NewReplacer(
		"http://", "",
		"https://", "",
		"/", "_",
		":", "_",
	).Replace(repoURL)
	if len(repoName) > 100 {
		repoName = repoName[:100]
	}

	clonePath := filepath.Join(g.tempDir, repoName)

	// Perform sparse clone
	args := []string{"clone", "--filter=blob:none", "--no-checkout"}
	args = append(args, repoURL, clonePath)
	cmd := exec.Command("git", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git clone failed: %w\nOutput: %s", err, output)
	}

	// Initialize sparse checkout
	cmd = exec.Command("git", "sparse-checkout", "init", "--cone")
	cmd.Dir = clonePath

	output, err = cmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(clonePath)

		return "", fmt.Errorf("git sparse-checkout init failed: %w\nOutput: %s", err, output)
	}

	g.ClonedRepos[repoURL] = clonePath

	return clonePath, nil
}

// Checks out a specific file from the cloned repository.
func (g *GitHelper) CheckoutFile(repoPath string, filePath string) error {
	// First, add the file to sparse-checkout
	dir := filepath.Dir(filePath)
	if dir != "." && dir != "" {
		cmd := exec.Command("git", "sparse-checkout", "add", dir)
		cmd.Dir = repoPath

		_, err := cmd.CombinedOutput()
		if err != nil {
			// Try to add the specific file if directory fails
			cmd = exec.Command("git", "sparse-checkout", "add", filePath)
			cmd.Dir = repoPath

			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("git sparse-checkout add failed: %w\nOutput: %s", err, output)
			}
		}
	}

	// Checkout the file
	cmd := exec.Command("git", "checkout", "HEAD", "--", filePath)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		// File might not exist in the repository
		return fmt.Errorf("git checkout failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// Checks if a file exists in a Git repository by attempting to check it out.
func (g *GitHelper) FileExistsInRepo(repoURL string, filePath string) (bool, string, error) {
	// Clone the repository if not already cloned
	clonePath, err := g.CloneRepo(repoURL)
	if err != nil {
		return false, "", err
	}

	// Try to checkout the file
	err = g.CheckoutFile(clonePath, filePath)
	if err != nil {
		// File doesn't exist in the repository
		return false, "", err
	}

	// File exists and has been checked out
	localPath := filepath.Join(clonePath, filePath)
	if _, err := os.Stat(localPath); err != nil {
		// File was supposedly checked out but doesn't exist - this is an error
		return false, "", fmt.Errorf("file was checked out but not found at %s: %w", localPath, err)
	}

	return true, localPath, nil
}

// Removes all temporary directories and cloned repositories.
func (g *GitHelper) Cleanup() error {
	if g.tempDir != "" {
		return os.RemoveAll(g.tempDir)
	}

	return nil
}

// Checks if an URL is a generic Git repository URL.
// Returns false for supported hosting platforms (GitHub, GitLab, Bitbucket)
// which have web interfaces and should not use local Git cloning.
func IsGitURL(u *url.URL) bool {
	if u == nil {
		return false
	}

	host := strings.ToLower(u.Host)
	switch host {
	case "github.com":
		return false
	case "gitlab.com":
		return false
	case "bitbucket.org":
		return false
	}

	if u.Scheme == "git" {
		return true
	}

	if u.Scheme == "http" || u.Scheme == "https" {
		if strings.HasSuffix(u.Path, ".git") ||
			strings.Contains(u.Path, ".git/") {
			return true
		}
	}

	return false
}

// Extracts the base repository URL in a generic Git repository.
func GetRepoURL(u *url.URL) string {
	repoURL := *u

	// For generic Git repos, remove the file path if present
	if idx := strings.Index(repoURL.Path, ".git/"); idx != -1 {
		// Keep everything up to and including .git
		repoURL.Path = repoURL.Path[:idx+4]

		return repoURL.String()
	}

	return repoURL.String()
}

// Extracts the file path from a generic Git repository URL.
func ExtractFilePathFromURL(u *url.URL) (string, error) {
	urlPath := u.Path

	var filePath string

	if strings.Contains(urlPath, ".git/") {
		if idx := strings.Index(urlPath, ".git/"); idx != -1 {
			filePath = urlPath[idx+5:]
		}
	}

	if filePath == "" {
		return "", fmt.Errorf("could not extract file path from URL: %s", u.String())
	}

	// Clean up the file path
	filePath = path.Clean(filePath)

	return filePath, nil
}
