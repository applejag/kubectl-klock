package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/replicatedhq/krew-plugin-template/pkg/logger"
)

type TemplateContext struct {
	Owner      string
	Repo       string
	PluginName string
}

func main() {
	owner, err := promptForOwner()
	if err != nil {
		fmt.Printf("%v\n", errors.Unwrap(err))
		os.Exit(1)
	}

	repo, err := promptForRepo()
	if err != nil {
		fmt.Printf("%v\n", errors.Unwrap(err))
		os.Exit(1)
	}

	pluginName, err := promptForPluginName()
	if err != nil {
		fmt.Printf("%v\n", errors.Unwrap(err))
		os.Exit(1)
	}

	templateContext := TemplateContext{
		Owner:      owner,
		Repo:       repo,
		PluginName: pluginName,
	}

	log := logger.NewLogger()
	log.Info("Updating README")
	if err := renderReadme(templateContext); err != nil {
		fmt.Printf("%v\n", errors.Unwrap(err))
		os.Exit(1)
	}

	log.Info("Updating sample code with names")
	if err := renderTemplates(templateContext); err != nil {
		fmt.Printf("%v\n", errors.Unwrap(err))
		os.Exit(1)
	}

	log.Info("Updating go.mod")
	if err := renderGoMod(templateContext); err != nil {
		fmt.Printf("%v\n", errors.Unwrap(err))
		os.Exit(1)
	}

	log.Info("Removing the setup application")
	if err := os.RemoveAll(path.Join("..", "setup")); err != nil {
		fmt.Printf("%v\n", errors.Unwrap(err))
		os.Exit(1)
	}

	log.Instructions("Your kubectl krew plugin has been initialized in this directory.")
	log.Instructions("For next steps, visit https://github.com/replicatedhq/krew-plugin-template")
	log.Info("")
}

func promptForOwner() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working dir: %w", err)
	}

	pathParts := strings.Split(filepath.Dir(cwd), string(os.PathSeparator))

	validate := func(input string) error {
		if len(input) == 0 {
			return errors.New("Invalid")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "GitHub Organization or Username",
		Validate: validate,
		Default:  pathParts[len(pathParts)-2],
	}

	result, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("failed to prompt for github owner: %w", err)
	}

	return result, nil
}

func promptForRepo() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working dir: %w", err)
	}

	pathParts := strings.Split(filepath.Dir(cwd), string(os.PathSeparator))

	validate := func(input string) error {
		if len(input) == 0 {
			return errors.New("invalid")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "GitHub Repo Name",
		Validate: validate,
		Default:  pathParts[len(pathParts)-1],
	}

	result, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("failed to prompt for github repo: %w", err)
	}

	return result, nil
}

func promptForPluginName() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working dir: %w", err)
	}

	pathParts := strings.Split(filepath.Dir(cwd), string(os.PathSeparator))

	validate := func(input string) error {
		if len(input) == 0 {
			return errors.New("Invalid")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Plugin Name",
		Validate: validate,
		Default:  pathParts[len(pathParts)-1],
	}

	result, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("failed to prompt for plugin name: %w", err)
	}

	return result, nil
}
