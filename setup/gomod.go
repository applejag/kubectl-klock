package main

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"
)

func renderGoMod(templateContext TemplateContext) error {
	input, err := ioutil.ReadFile(path.Join("..", "go.mod"))
	if err != nil {
		return fmt.Errorf("failed to read go.mod: %w", err)
	}

	lines := strings.Split(string(input), "\n")

	lines[0] = fmt.Sprintf(`module github.com/%s/%s`, templateContext.Owner, templateContext.Repo)

	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(path.Join("..", "go.mod"), []byte(output), 0644)
	if err != nil {
		return fmt.Errorf("failed to write go.mod: %w", err)
	}

	return nil
}
