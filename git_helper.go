package publiccode

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
type gitHelper struct {
	// Base directory for all temporary Git clones
	tempDir string
	// Maps repository URLs to their local clone paths
	clonedRepos map[string]string
}

func newGitHelper() (*gitHelper, error) {
	tempDir, err := os.MkdirTemp("", "publiccode-git-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return &gitHelper{
		tempDir:     tempDir,
		clonedRepos: make(map[string]string),
	}, nil
}

// Performs a sparse clone of a Git repository.
func (g *gitHelper) cloneRepo(repoURL string) (string, error) {
	// Check if already cloned
	if clonePath, ok := g.clonedRepos[repoURL]; ok {
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

	g.clonedRepos[repoURL] = clonePath
	return clonePath, nil
}

// Checks out a specific file from the cloned repository.
func (g *gitHelper) checkoutFile(repoPath string, filePath string) error {
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
func (g *gitHelper) fileExistsInRepo(repoURL string, filePath string) (bool, string, error) {
	// Clone the repository if not already cloned
	clonePath, err := g.cloneRepo(repoURL)
	if err != nil {
		return false, "", err
	}

	// Try to checkout the file
	err = g.checkoutFile(clonePath, filePath)
	if err != nil {
		// File doesn't exist in the repository
		return false, "", nil
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
func (g *gitHelper) cleanup() error {
	if g.tempDir != "" {
		return os.RemoveAll(g.tempDir)
	}
	return nil
}

// Checks if an URL is a generic Git repository URL.
// Returns false for supported hosting platforms (GitHub, GitLab, Bitbucket)
// which have web interfaces and should not use local Git cloning.
func isGitURL(u *url.URL) bool {
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
func getRepoURL(u *url.URL) string {
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
func extractFilePathFromURL(u *url.URL) (string, error) {
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

// Checks if a file exists in a Git repository.
func (p *Parser) checkFileInGitRepo(u *url.URL) (bool, string, error) {
	if !p.allowLocalGitClone {
		return false, "", fmt.Errorf("local Git clone not allowed")
	}

	// Extract repository URL and file path
	repoURL := getRepoURL(u)
	filePath, err := extractFilePathFromURL(u)
	if err != nil {
		return false, "", fmt.Errorf("failed to extract file path from URL %s: %w", u.String(), err)
	}

	if p.gitRepoCache == nil {
		p.gitRepoCache = make(map[string]string)
	}
	if cachedPath, ok := p.gitRepoCache[repoURL]; ok {
		// Check if file exists in cached repo
		localPath := filepath.Join(cachedPath, filePath)
		if _, err := os.Stat(localPath); err == nil {
			return true, localPath, nil
		}
	}

	// Create a temporary Git helper
	helper, err := newGitHelper()
	if err != nil {
		return false, "", fmt.Errorf("failed to create Git helper: %w", err)
	}
	// Don't cleanup the temp directory, will be managed by the Parser

	// Check if file exists in repository
	exists, localPath, err := helper.fileExistsInRepo(repoURL, filePath)
	if err != nil {
		helper.cleanup() // Clean up on error
		return false, "", fmt.Errorf("failed to check file in repo %s: %w", repoURL, err)
	}

	if exists {
		// Cache the cloned repo path for future use
		// The path will be cleaned up when Parser.Cleanup() is called
		if len(helper.clonedRepos) > 0 {
			for _, clonePath := range helper.clonedRepos {
				p.gitRepoCache[repoURL] = clonePath
				break
			}
		}
	} else {
		// If file doesn't exist, clean up immediately
		helper.cleanup()
	}

	return exists, localPath, nil
}
