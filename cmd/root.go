package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jilleJr/kubectl-klock/pkg/klock"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func RootCmd() *cobra.Command {
	var o klock.Options
	cmd := &cobra.Command{
		Use:           "klock",
		Short:         "",
		Long:          `.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.MinimumNArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := klock.Execute(o, args); err != nil {
				return err
			}
			return nil
		},
	}

	cobra.OnInitialize(initConfig)

	o.ConfigFlags = genericclioptions.NewConfigFlags(false)
	o.ConfigFlags.AddFlags(cmd.Flags())

	cmd.Flags().BoolVarP(&o.AllNamespaces, "all-namespaces", "A", o.AllNamespaces, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	cmd.Flags().StringVar(&o.FieldSelector, "field-selector", o.FieldSelector, "Selector (field query) to filter on, supports '=', '==', and '!='.(e.g. --field-selector key1=value1,key2=value2). The server only supports a limited number of field queries per type.")
	cmd.Flags().BoolVar(&o.OutputWatchEvents, "output-watch-events", o.OutputWatchEvents, "Output watch event objects when --watch or --watch-only is used. Existing objects are output as initial ADDED events.")
	cmdutil.AddLabelSelectorFlagVar(cmd, &o.LabelSelector)

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
