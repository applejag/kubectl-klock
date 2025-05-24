// SPDX-FileCopyrightText: 2023 Kalle Fagerberg
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the
// Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
// more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/kubecolor/kubecolor/config"
	"github.com/kubecolor/kubecolor/printer"
	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/completion"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/applejag/kubectl-klock/pkg/klock"
	"github.com/applejag/kubectl-klock/pkg/types"
)

var (
	Stdout = colorable.NewColorableStdout()
	Stderr = colorable.NewColorableStderr()
)

var Version string

func RootCmd(kubecolorConfig *config.Config) *cobra.Command {
	kubeConfigFlags := genericclioptions.NewConfigFlags(false)
	f := cmdutil.NewFactory(kubeConfigFlags)

	use := "kubectl-klock"
	if useEnv := os.Getenv("KLOCK_USAGE_NAME"); useEnv != "" {
		use = useEnv
	}

	k := initConfig()

	var o klock.Options
	cmd := &cobra.Command{
		Use:   use,
		Short: "Watches resources",
		Long: `Watches resources.

 Prints a table of the most important information about the specified resource.

 Supports a lot of the flags that regular "kubectl get" supports,
 such as using the label selector (--selector, -l),
 show results of all namespaces (--all-namespaces, -A),
 as well as settings output format (--output, -o).

 Performs the equivalent to running "watch kubectl get pods", but using
 the same protocol as "kubectl get pods --watch".

Use "kubectl api-resources" for a complete list of supported resources.`,
		Example: `  # Watch all pods
  kubectl klock pods

  # Watch all pods with more information (such as node name)
  kubectl klock pods -o wide

  # Watch a specific pod
  kubectl klock pods my-pod-7d68885db5-6dfst

  # Watch a subset of pods, filtering on labels
  kubectl klock pods --selector app=my-app
  kubectl klock pods -l app=my-app

  # Watch all pods in all namespaces
  kubectl klock pods --all-namespaces
  kubectl klock pods -A

  # Watch other resource types
  kubectl klock cronjobs
  kubectl klock deployments
  kubectl klock statefulsets
  kubectl klock nodes

  # Watch all pods, but restart the watch when your ~/.kube/config file changes,
  # such as when using "kubectl config use-context NAME"
  kubectl klock pods --watch-kubeconfig
  kubectl klock pods -W`,
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := k.Load(posflag.Provider(cmd.Flags(), ".", k), nil); err != nil {
				return err
			}

			if err := k.Unmarshal("", &o); err != nil {
				return err
			}

			return nil
		},
		Version: Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return klock.Execute(o, args)
		},
		ValidArgsFunction: completion.ResourceTypeAndNameCompletionFunc(f),
	}

	o.Kubecolor = kubecolorConfig
	o.HideDeleted = types.NewOptionalDuration(30 * time.Second)

	o.ConfigFlags = kubeConfigFlags
	o.ConfigFlags.AddFlags(cmd.Flags())

	cmd.Flags().BoolP("all-namespaces", "A", o.AllNamespaces, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	cmd.Flags().String("field-selector", o.FieldSelector, "Selector (field query) to filter on, supports '=', '==', and '!='.(e.g. --field-selector key1=value1,key2=value2). The server only supports a limited number of field queries per type.")
	cmd.Flags().StringP("output", "o", o.Output, "Output format. Only a small subset of formats found in 'kubectl get' are supported by kubectl-klock.")
	cmd.Flags().BoolP("watch-kubeconfig", "W", o.WatchKubeconfig, "Restart the watch when the kubeconfig file changes.")
	cmd.Flags().StringSliceP("label-columns", "L", o.LabelColumns, "Accepts a comma separated list of labels that are going to be presented as columns.")
	cmd.Flags().Var(&o.HideDeleted, "hide-deleted", `Hide deleted elements after this duration. Example: "10s", "1m". Set to "0" to always hide, and "false" to show forever.`)
	cmdutil.AddLabelSelectorFlagVar(cmd, &o.LabelSelector)

	cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"wide"}, cobra.ShellCompDirectiveNoFileComp
	})

	registerCompletionFuncForGlobalFlags(cmd, f)

	// Must add temporary subcommand, as Cobra won't add completion commands
	// if the command doesn't have any subcommands.
	tmpChild := &cobra.Command{Use: "tmp", Hidden: true}
	cmd.AddCommand(tmpChild)
	cmd.InitDefaultCompletionCmd()
	cmd.RemoveCommand(tmpChild)

	cmd.SetErr(Stderr)
	cmd.SetOut(Stdout)

	// Use kubectl --help templates

	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		templates.ActsAsRootCommand(cmd, nil, templates.CommandGroup{})
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.Help()
		p := printer.HelpPrinter{
			Theme: &kubecolorConfig.Theme,
		}
		p.Print(&buf, Stdout)
	})

	return cmd
}

func InitAndExecute() {
	kubecolorConfig, err := getKubecolorConfig()
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
	if err := RootCmd(kubecolorConfig).Execute(); err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
}

func initConfig() *koanf.Koanf {
	k := koanf.New(".")

	replacer := strings.NewReplacer("_", "-")
	k.Load(env.Provider("KLOCK_", "__", func(s string) string {
		return replacer.Replace(
			strings.ToLower(
				strings.TrimPrefix(s, "KLOCK_"),
			),
		)
	}), nil)

	return k
}

func getKubecolorConfig() (*config.Config, error) {
	v, err := config.LoadViper()
	if err != nil {
		return nil, err
	}
	if err := config.ApplyThemePreset(v); err != nil {
		return nil, err
	}
	return config.Unmarshal(v)
}
