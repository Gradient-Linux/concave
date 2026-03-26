package cmd

func progressImageLabel(image string) string {
	if len(image) <= 36 {
		return image
	}
	return image[:33] + "..."
}
