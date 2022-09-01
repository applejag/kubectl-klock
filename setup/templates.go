package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func renderTemplates(templateContext TemplateContext) error {
	err := filepath.Walk("..",
		func(path string, info os.FileInfo, err error) error {
			// ignore .git
			pathSplit := strings.Split(path, string(os.PathSeparator))
			if len(pathSplit) >= 2 {
				if pathSplit[1] == ".git" {
					return nil
				}
			}

			// ignore the setup pkg
			if len(pathSplit) >= 2 {
				if pathSplit[1] == "setup" {
					return nil
				}
			}

			// Ignore bin
			if len(pathSplit) >= 2 {
				if pathSplit[1] == "bin" {
					return nil
				}
			}

			// Ignore .github
			if len(pathSplit) >= 2 {
				if pathSplit[1] == ".github" {
					return nil
				}
			}

			fi, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("failed to read file info: %w", err)
			}

			if fi.IsDir() {
				return nil
			}

			tmpl, err := template.ParseFiles(path)
			if err != nil {
				return fmt.Errorf("failed to parse template: %w", err)
			}

			f, err := os.Create(path)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}

			if err := tmpl.Execute(f, templateContext); err != nil {
				return fmt.Errorf("failed to execute template: %w", err)
			}

			return nil
		})
	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	return nil
}
