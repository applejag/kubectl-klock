package main

import (
	"fmt"
	"io/ioutil"
	"path"
)

func renderReadme(templateContext TemplateContext) error {
	input, err := ioutil.ReadFile(path.Join("hack", "README.txt"))
	if err != nil {
		return fmt.Errorf("failed to read README.txt: %w", err)
	}

	err = ioutil.WriteFile(path.Join("..", "README.md"), input, 0644)
	if err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	return nil
}
