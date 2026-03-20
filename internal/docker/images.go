package docker

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// PullWithProgress tags an existing image as previous and then pulls the target image.
func PullWithProgress(ctx context.Context, image string, cb func(string)) error {
	if err := TagAsPrevious(image); err != nil {
		return err
	}
	return Pull(ctx, image, cb)
}

// TagAsPrevious tags an existing image as <repo>:gradient-previous when present.
func TagAsPrevious(image string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err := runCombinedOutput(ctx, "docker", "image", "inspect", image); err != nil {
		return nil
	}
	previous := previousImageTag(image)
	if _, err := runCombinedOutput(ctx, "docker", "tag", image, previous); err != nil {
		return fmt.Errorf("docker tag %s %s: %w", image, previous, err)
	}
	return nil
}

// RevertToPrevious retags <repo>:gradient-previous back to the requested image tag.
func RevertToPrevious(image string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	previous := previousImageTag(image)
	if _, err := runCombinedOutput(ctx, "docker", "image", "inspect", previous); err != nil {
		return fmt.Errorf("inspect previous image %s: %w", previous, err)
	}
	if _, err := runCombinedOutput(ctx, "docker", "tag", previous, image); err != nil {
		return fmt.Errorf("docker tag %s %s: %w", previous, image, err)
	}
	return nil
}

func previousImageTag(image string) string {
	lastSlash := strings.LastIndex(image, "/")
	lastColon := strings.LastIndex(image, ":")
	if lastColon > lastSlash {
		return image[:lastColon] + ":gradient-previous"
	}
	return image + ":gradient-previous"
}
