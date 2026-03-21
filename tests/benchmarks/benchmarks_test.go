package benchmarks

import (
	"testing"

	"github.com/Gradient-Linux/concave/internal/config"
	"github.com/Gradient-Linux/concave/internal/suite"
	"github.com/Gradient-Linux/concave/internal/workspace"
)

type benchmarkRecord struct {
	name   string
	images map[string]string
}

func (b benchmarkRecord) RecordName() string {
	return b.name
}

func (b benchmarkRecord) RecordImages() map[string]string {
	return b.images
}

func BenchmarkSuiteLookup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := suite.Get("boosting"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkManifestRoundTrip(b *testing.B) {
	b.StopTimer()
	b.Setenv("HOME", b.TempDir())
	_ = workspace.EnsureLayout()
	record := benchmarkRecord{
		name: "boosting",
		images: map[string]string{
			"gradient-boost-core": "python:3.12-slim",
		},
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		manifest := config.RecordInstall(config.VersionManifest{}, record)
		manifest = config.RecordUpdate(manifest, "boosting", "gradient-boost-core", "python:3.12-alpine")
		if err := config.SaveManifest(manifest); err != nil {
			b.Fatal(err)
		}
		if _, err := config.LoadManifest(); err != nil {
			b.Fatal(err)
		}
	}
}
