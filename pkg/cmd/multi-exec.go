package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	tcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
)

type Executor struct {
	CoreV1Client tcorev1.CoreV1Interface
	Config       *restclient.Config
	ctx          context.Context
	genericclioptions.IOStreams
}

func Execute(matchVersionFlags *cmdutil.MatchVersionFlags) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		f := cmdutil.NewFactory(matchVersionFlags)

		if len(args) == 0 {
			log.Fatal("error: you must specify at least one command for the container")
		}

		streams := genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		}

		clientConfig, err := f.ToRESTConfig()
		if err != nil {
			log.Fatal(err)
		}

		coreClient, err := corev1client.NewForConfig(clientConfig)
		if err != nil {
			log.Fatal(err)
		}

		selector, err := cmd.Flags().GetString("selector")
		if err != nil {
			log.Fatal(err)
		}

		namespace, err := cmd.Flags().GetString("namespace")
		if err != nil {
			log.Fatal(err)
		}

		executor := &Executor{
			CoreV1Client: coreClient,
			Config:       clientConfig,
			ctx:          context.TODO(),
			IOStreams:    streams,
		}

		err = executor.executeInPod(selector, namespace, args)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (executor *Executor) executeInPod(selector, namespace string, command []string) error {
	podList, err := executor.CoreV1Client.Pods(namespace).List(executor.ctx, metav1.ListOptions{
		LabelSelector: selector,
	})

	if err != nil {
		return err
	}

	if len(podList.Items) == 0 {
		return fmt.Errorf("No resource found in %s namespace, selector %s", namespace, selector)
	}

	pods := podList.Items
	for _, pod := range pods {
		req := executor.CoreV1Client.RESTClient().Post().Resource("pods").Name(pod.Name).Namespace("lab").SubResource("exec")
		option := &v1.PodExecOptions{
			Command: command,
			Stdin:   true,
			Stdout:  true,
			Stderr:  true,
			TTY:     true,
		}

		req.VersionedParams(
			option,
			scheme.ParameterCodec,
		)

		exec, err := remotecommand.NewSPDYExecutor(executor.Config, "POST", req.URL())
		if err != nil {
			return err
		}

		err = exec.Stream(remotecommand.StreamOptions{
			Stdin:  executor.IOStreams.In,
			Stdout: executor.IOStreams.Out,
			Stderr: executor.IOStreams.ErrOut,
		})

		if err != nil {
			return err
		}
	}
	return nil
}
