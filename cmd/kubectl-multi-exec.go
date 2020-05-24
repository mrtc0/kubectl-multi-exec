package main

import (
	"log"

	"github.com/mrtc0/kubectl-multi-exec/pkg/cmd"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func main() {
	command := &cobra.Command{
		Use: "kubectl-multi-exec",
	}

	configFlags := genericclioptions.NewConfigFlags(false)
	configFlags.AddFlags(command.PersistentFlags())

	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	matchVersionFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	matchVersionFlags.AddFlags(command.PersistentFlags())

	command.PersistentFlags().String(
		"selector",
		"key1=value1",
		"Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)",
	)

	command.Run = cmd.Execute(matchVersionFlags)
	err := command.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
