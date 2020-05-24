package main

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

type Injector struct {
	ctx context.Context
	genericclioptions.IOStreams
	CoreV1Client tcorev1.CoreV1Interface
	Config       *restclient.Config
}

func main() {
	cmds := &cobra.Command{
		Use: "multi-exec",
	}

	configFlags := genericclioptions.NewConfigFlags(false)
	configFlags.AddFlags(cmds.PersistentFlags())

	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	matchVersionKubeConfigFlags.AddFlags(cmds.PersistentFlags())

	cmds.PersistentFlags().String("selector", "app=myapp", "a kubernetes label selector to choose the pods to run command")

	cmds.Run = runCmd(matchVersionKubeConfigFlags)
	err := cmds.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func runCmd(matchVersionKubeConfigFlags *cmdutil.MatchVersionFlags) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

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

		injector := &Injector{
			CoreV1Client: coreClient,
			Config:       clientConfig,
			ctx:          context.TODO(),
			IOStreams:    streams,
		}

		err = injector.executeInPod(selector, namespace, args)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (injector *Injector) executeInPod(selector, namespace string, command []string) error {
	podList, err := injector.CoreV1Client.Pods(namespace).List(injector.ctx, metav1.ListOptions{
		LabelSelector: selector,
	})

	if err != nil {
		return err
	}

	if len(podList.Items) == 0 {
		return fmt.Errorf("no pod found to attach with the given selector: %s", selector)
	}

	pods := podList.Items
	for _, pod := range pods {
		req := injector.CoreV1Client.RESTClient().Post().Resource("pods").Name(pod.Name).Namespace("lab").SubResource("exec")
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

		exec, err := remotecommand.NewSPDYExecutor(injector.Config, "POST", req.URL())
		if err != nil {
			return err
		}

		err = exec.Stream(remotecommand.StreamOptions{
			Stdin:  injector.IOStreams.In,
			Stdout: injector.IOStreams.Out,
			Stderr: injector.IOStreams.ErrOut,
		})

		if err != nil {
			return err
		}
	}
	return nil
}
