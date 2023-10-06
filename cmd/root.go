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
	"fmt"
	"os"
	"strings"

	"github.com/applejag/kubectl-klock/pkg/klock"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/completion"
)

var Version string

func RootCmd() *cobra.Command {
	kubeConfigFlags := genericclioptions.NewConfigFlags(false)
	f := cmdutil.NewFactory(kubeConfigFlags)

	use := "klock"
	if useEnv := os.Getenv("KLOCK_USAGE_NAME"); useEnv != "" {
		use = useEnv
	}

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

Use "kubectl api-resources" for a complete list of supported resources.

Examples:
  # Watch all pods
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
  kubectl klock pods -W
`,
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.MinimumNArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		Version: Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return klock.Execute(o, args)
		},
		ValidArgsFunction: completion.ResourceTypeAndNameCompletionFunc(f),
	}

	cobra.OnInitialize(initConfig)

	o.ConfigFlags = kubeConfigFlags
	o.ConfigFlags.AddFlags(cmd.Flags())

	cmd.Flags().BoolVarP(&o.AllNamespaces, "all-namespaces", "A", o.AllNamespaces, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	cmd.Flags().StringVar(&o.FieldSelector, "field-selector", o.FieldSelector, "Selector (field query) to filter on, supports '=', '==', and '!='.(e.g. --field-selector key1=value1,key2=value2). The server only supports a limited number of field queries per type.")
	cmd.Flags().StringVarP(&o.Output, "output", "o", o.Output, "Output format. Only a small subset of formats found in 'kubectl get' are supported by kubectl-klock.")
	cmd.Flags().BoolVarP(&o.WatchKubeconfig, "watch-kubeconfig", "W", o.WatchKubeconfig, "Restart the watch when the kubeconfig file changes.")
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

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	return cmd
}

func InitAndExecute() {
	if err := RootCmd().Execute(); err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
}

func initConfig() {
	viper.AutomaticEnv()
}
