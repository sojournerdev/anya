package replay

import (
	"embed"
	"fmt"
	"io"
)

//go:embed testdata/*.jsonl
var embeddedFixtures embed.FS

func openEmbeddedFixture(name string) (io.ReadCloser, error) {
	path, err := embeddedFixturePath(name)
	if err != nil {
		return nil, err
	}
	file, err := embeddedFixtures.Open(path)
	if err != nil {
		return nil, fmt.Errorf("replay: open embedded fixture %q: %w", name, err)
	}
	return file, nil
}

func embeddedFixturePath(name string) (string, error) {
	switch name {
	case "healthy":
		return "testdata/healthy.jsonl", nil
	default:
		return "", fmt.Errorf("replay: unknown embedded fixture %q", name)
	}
}
