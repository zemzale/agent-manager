package git

import (
	"path"
	"strings"
)

// ExtractRepoName extracts the repository name from a git URL.
// Handles HTTPS (https://github.com/user/repo.git) and SSH (git@github.com:user/repo.git) formats.
func ExtractRepoName(url string) string {
	// Handle SSH format: git@github.com:user/repo.git
	if idx := strings.LastIndex(url, ":"); idx != -1 && !strings.Contains(url, "://") {
		url = url[idx+1:]
	}
	name := path.Base(url)
	return strings.TrimSuffix(name, ".git")
}
