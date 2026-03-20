package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gradient-linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

const defaultManifestURL = "https://packages.gradientlinux.dev/concave/latest.json"

type updateManifest struct {
	Version string `json:"version"`
	URL     string `json:"url"`
	SHA256  string `json:"sha256"`
}

var selfUpdateCmd = &cobra.Command{
	Use:   "self-update",
	Short: "Download and atomically replace the concave binary",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := http.Get(defaultManifestURL)
		if err != nil {
			return fmt.Errorf("download manifest: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("manifest returned %s", resp.Status)
		}

		var manifest updateManifest
		if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
			return fmt.Errorf("decode manifest: %w", err)
		}

		binResp, err := http.Get(manifest.URL)
		if err != nil {
			return fmt.Errorf("download binary: %w", err)
		}
		defer binResp.Body.Close()
		if binResp.StatusCode != http.StatusOK {
			return fmt.Errorf("binary returned %s", binResp.Status)
		}

		tempPath := filepath.Join(os.TempDir(), "concave.new")
		file, err := os.Create(tempPath)
		if err != nil {
			return fmt.Errorf("create %s: %w", tempPath, err)
		}

		hasher := sha256.New()
		writer := io.MultiWriter(file, hasher)
		if _, err := io.Copy(writer, binResp.Body); err != nil {
			file.Close()
			return fmt.Errorf("write %s: %w", tempPath, err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("close %s: %w", tempPath, err)
		}

		sum := hex.EncodeToString(hasher.Sum(nil))
		if sum != manifest.SHA256 {
			return fmt.Errorf("sha256 mismatch: expected %s got %s", manifest.SHA256, sum)
		}
		if err := os.Chmod(tempPath, 0o755); err != nil {
			return fmt.Errorf("chmod %s: %w", tempPath, err)
		}
		if err := os.Rename(tempPath, "/usr/local/bin/concave"); err != nil {
			return fmt.Errorf("replace /usr/local/bin/concave: %w", err)
		}
		ui.Pass("Updated", manifest.Version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(selfUpdateCmd)
}
