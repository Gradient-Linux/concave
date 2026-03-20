package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/gradientlinux/concave/internal/config"
	"github.com/gradientlinux/concave/internal/suite"
	"github.com/gradientlinux/concave/internal/system"
	"github.com/spf13/cobra"
)

var labSuite string

var labCmd = &cobra.Command{
	Use:   "lab",
	Short: "Open JupyterLab for an installed suite",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := chooseLabSuite()
		if err != nil {
			return err
		}
		if labSuite != "" {
			name = labSuite
		}

		s, err := suite.Get(name)
		if err != nil {
			return err
		}
		container, ok := suite.JupyterContainer(s)
		if !ok {
			return fmt.Errorf("suite %s has no JupyterLab service", s.Name)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		out, err := exec.CommandContext(ctx, "docker", "exec", container, "jupyter", "server", "list").CombinedOutput()
		if err != nil {
			out, err = exec.CommandContext(ctx, "docker", "logs", container).CombinedOutput()
			if err != nil {
				return fmt.Errorf("resolve Jupyter token for %s: %w", container, err)
			}
		}

		url, err := extractLabURL(string(out))
		if err != nil {
			return err
		}
		return system.OpenURL(url)
	},
}

func chooseLabSuite() (string, error) {
	state, err := config.LoadState()
	if err != nil {
		return "", err
	}
	for _, name := range state.Installed {
		s, err := suite.Get(name)
		if err != nil {
			return "", err
		}
		if _, ok := suite.JupyterContainer(s); ok {
			return name, nil
		}
	}
	return "", fmt.Errorf("no installed suite exposes JupyterLab")
}

func extractLabURL(raw string) (string, error) {
	re := regexp.MustCompile(`https?://[^\s]+/\??[^\s]*token=[A-Za-z0-9]+`)
	match := re.FindString(raw)
	if match == "" {
		return "", fmt.Errorf("unable to find tokenized Jupyter URL")
	}
	match = strings.Replace(match, "0.0.0.0", "127.0.0.1", 1)
	match = strings.Replace(match, "localhost", "127.0.0.1", 1)
	match = strings.Replace(match, "/?token=", "/lab?token=", 1)
	return match, nil
}

func init() {
	labCmd.Flags().StringVar(&labSuite, "suite", "", "suite to open explicitly")
	rootCmd.AddCommand(labCmd)
}
