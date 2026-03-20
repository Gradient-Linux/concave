package benchmarks

import (
	"testing"

	"github.com/gradient-linux/concave/internal/config"
	"github.com/gradient-linux/concave/internal/docker"
	"github.com/gradient-linux/concave/internal/suite"
	"github.com/gradient-linux/concave/internal/workspace"
)

func BenchmarkSuiteLookup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := suite.Get("boosting"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRenderSuiteCompose(b *testing.B) {
	b.StopTimer()
	b.Setenv("HOME", b.TempDir())
	_ = workspace.EnsureLayout()
	s, err := suite.Get("boosting")
	if err != nil {
		b.Fatal(err)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		if _, err := docker.RenderSuiteCompose(s); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVersionsRoundTrip(b *testing.B) {
	b.StopTimer()
	b.Setenv("HOME", b.TempDir())
	_ = workspace.EnsureLayout()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		versions := config.Versions{}
		config.SetImageVersion(versions, "boosting", "gradient-boost-core", "python:3.12-slim", "")
		if err := config.SaveVersions(versions); err != nil {
			b.Fatal(err)
		}
		if _, err := config.LoadVersions(); err != nil {
			b.Fatal(err)
		}
	}
}
