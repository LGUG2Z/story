package manifest

import (
	"bufio"
	"fmt"
	"os"
)

func LoadStoryIgnore(basePath string) (map[string]bool, error) {
	storyIgnoreFilePath := fmt.Sprintf("%s/.storyignore", basePath)
	storyIgnore := make(map[string]bool)

	if _, err := os.Stat(storyIgnoreFilePath); os.IsNotExist(err) {
		return storyIgnore, nil
	}

	file, err := os.Open(storyIgnoreFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		storyIgnore[scanner.Text()] = true
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return storyIgnore, nil
}
