package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var labSuite string

var labCmd = &cobra.Command{
	Use:   "lab",
	Short: "Open JupyterLab for an installed suite",
	RunE:  runLab,
}

func runLab(cmd *cobra.Command, args []string) error {
	name, err := chooseLabSuite()
	if err != nil {
		return err
	}
	if labSuite != "" {
		name = labSuite
	}

	installed, err := isInstalled(name)
	if err != nil {
		return err
	}
	if !installed {
		return fmt.Errorf("suite %s is not installed", name)
	}

	s, err := currentSuiteDefinition(name)
	if err != nil {
		return err
	}
	container, ok := jupyterContainer(s)
	if !ok {
		return fmt.Errorf("suite %s has no JupyterLab service", s.Name)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	composePath := dockerComposePath(name)
	status, err := dockerComposeServiceStatus(ctx, composePath, container)
	if err != nil {
		return err
	}
	if status != "running" {
		return fmt.Errorf("JupyterLab is not running. Start it with: concave start %s", name)
	}

	out, err := dockerComposeExecOutput(ctx, composePath, container, "jupyter", "server", "list", "--json")
	if err != nil {
		return fmt.Errorf("resolve Jupyter token for %s: %w", container, err)
	}

	if url, ok := preferredGradientLabURL(); ok {
		ui.Info("gradient-lab", "opening at "+url)
		return systemOpenURL(url)
	}

	url, err := extractLabURL(string(out))
	if err != nil {
		return err
	}
	ui.Info("JupyterLab", "opening at "+url)
	return systemOpenURL(url)
}

func chooseLabSuite() (string, error) {
	if labSuite != "" {
		return labSuite, nil
	}

	state, err := loadState()
	if err != nil {
		return "", err
	}

	installed := make(map[string]struct{}, len(state.Installed))
	for _, name := range state.Installed {
		installed[name] = struct{}{}
	}
	for _, candidate := range []string{"boosting", "neural"} {
		if _, ok := installed[candidate]; ok {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("no installed suite exposes JupyterLab")
}

func extractLabURL(raw string) (string, error) {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var server struct {
			URL   string `json:"url"`
			Token string `json:"token"`
		}
		if err := json.Unmarshal([]byte(line), &server); err == nil && server.Token != "" {
			return "http://localhost:8888/lab?token=" + server.Token, nil
		}
	}

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

func preferredGradientLabURL() (string, bool) {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:8889", 750*time.Millisecond)
	if err != nil {
		return "", false
	}
	_ = conn.Close()
	return "http://127.0.0.1:8889/lab", true
}

func init() {
	labCmd.Flags().StringVar(&labSuite, "suite", "", "suite to open explicitly")
	rootCmd.AddCommand(labCmd)
}
