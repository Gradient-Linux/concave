package templatepath

import (
	"os"
	"path/filepath"
)

var executablePath = os.Executable

// Candidates returns likely runtime locations for a packaged or source-tree template file.
func Candidates(filename, callerFile string) []string {
	candidates := []string{filepath.Join("templates", filename)}

	if callerFile != "" {
		repoRoot := filepath.Clean(filepath.Join(filepath.Dir(callerFile), "..", ".."))
		candidates = append(candidates, filepath.Join(repoRoot, "templates", filename))
	}

	if executable, err := executablePath(); err == nil {
		exeDir := filepath.Dir(executable)
		candidates = append(candidates,
			filepath.Join(exeDir, "templates", filename),
			filepath.Clean(filepath.Join(exeDir, "..", "share", "concave", "templates", filename)),
		)
	}

	candidates = append(candidates,
		filepath.Join(string(filepath.Separator), "usr", "local", "share", "concave", "templates", filename),
		filepath.Join(string(filepath.Separator), "usr", "share", "concave", "templates", filename),
	)

	return unique(candidates)
}

func unique(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}
