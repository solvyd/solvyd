package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/vrenjith/ritmo/plugin-sdk/pkg/sdk"
)

// GitSCMPlugin implements SCM plugin for Git
type GitSCMPlugin struct {
	depth       int
	submodules  bool
	credentials string
}

func (p *GitSCMPlugin) Name() string {
	return "git-scm"
}

func (p *GitSCMPlugin) Version() string {
	return "1.0.0"
}

func (p *GitSCMPlugin) Type() string {
	return "scm"
}

func (p *GitSCMPlugin) Initialize(config map[string]interface{}) error {
	if depth, ok := config["depth"].(float64); ok {
		p.depth = int(depth)
	} else {
		p.depth = 0 // full clone
	}

	if submodules, ok := config["submodules"].(bool); ok {
		p.submodules = submodules
	}

	if creds, ok := config["credentials"].(string); ok {
		p.credentials = creds
	}

	return nil
}

func (p *GitSCMPlugin) Execute(ctx *sdk.ExecutionContext) (*sdk.Result, error) {
	ctx.Logger.Info("Starting Git clone operation")

	// Get repository URL from parameters
	url, ok := ctx.Parameters["url"].(string)
	if !ok {
		return &sdk.Result{
			Success:      false,
			ErrorMessage: "Repository URL not provided",
		}, fmt.Errorf("missing repository URL")
	}

	branch := "main"
	if b, ok := ctx.Parameters["branch"].(string); ok {
		branch = b
	}

	// Clone the repository
	if err := p.Clone(url, branch, "", ctx.WorkDir); err != nil {
		return &sdk.Result{
			Success:      false,
			ErrorMessage: err.Error(),
		}, err
	}

	ctx.Logger.Info("Git clone completed successfully")

	return &sdk.Result{
		Success:  true,
		ExitCode: 0,
		Output:   fmt.Sprintf("Cloned %s (branch: %s)", url, branch),
	}, nil
}

func (p *GitSCMPlugin) Clone(url, branch, commitSHA, dest string) error {
	args := []string{"clone"}

	if p.depth > 0 {
		args = append(args, "--depth", fmt.Sprintf("%d", p.depth))
	}

	if !p.submodules {
		args = append(args, "--no-recurse-submodules")
	}

	args = append(args, "--branch", branch, url, dest)

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	// Checkout specific commit if provided
	if commitSHA != "" {
		cmd := exec.Command("git", "checkout", commitSHA)
		cmd.Dir = dest
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git checkout failed: %w", err)
		}
	}

	return nil
}

func (p *GitSCMPlugin) GetCommitInfo(commitSHA string) (*sdk.CommitInfo, error) {
	// TODO: Implement commit info retrieval
	return &sdk.CommitInfo{
		SHA: commitSHA,
	}, nil
}

func (p *GitSCMPlugin) Cleanup() error {
	return nil
}

// Export the plugin
var Plugin GitSCMPlugin

// Main function for standalone testing
func main() {
	fmt.Println("Git SCM Plugin v1.0.0")
	fmt.Println("This is a plugin and should be loaded by the Ritmo system")
	fmt.Println("To build as a plugin: go build -buildmode=plugin -o git-scm.so")
}
