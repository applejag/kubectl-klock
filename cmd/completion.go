// SPDX-FileCopyrightText: 2019 The Kubernetes Authors.
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/completion"
)

// registerCompletionFuncForGlobalFlags was copied from
// [https://github.com/kubernetes/kubectl/blob/7ce6599fc128a522ffa08ba0a594be801afb5cd4/pkg/cmd/cmd.go#L548-L569]
func registerCompletionFuncForGlobalFlags(cmd *cobra.Command, f cmdutil.Factory) {
	completion.SetFactoryForCompletion(f)

	cmdutil.CheckErr(cmd.RegisterFlagCompletionFunc("namespace", completion.ResourceNameCompletionFunc(f, "namespace")))
	cmdutil.CheckErr(cmd.RegisterFlagCompletionFunc("context", completion.ContextCompletionFunc))
	cmdutil.CheckErr(cmd.RegisterFlagCompletionFunc("cluster", completion.ClusterCompletionFunc))
	cmdutil.CheckErr(cmd.RegisterFlagCompletionFunc("user", completion.UserCompletionFunc))
}
