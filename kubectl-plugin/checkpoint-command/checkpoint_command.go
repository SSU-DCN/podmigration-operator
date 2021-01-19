package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

/* To use this kubectl-plugin
 		$ go build -o kubectl-checkpoint
		$ sudo cp kubectl-checkpoint /usr/local/bin
*/

const (
	example = `
	# Checkpoint a running Pod and save the checkpoint infomations to given path
	kubectl checkpoint [POD_NAME] [CHECKPOINT_PATH]
	kubectl checkpoint [POD_NAME] --namespace string [CHECKPOINT_PATH]
	`
	longDesc = `
	checkpoint [POD_NAME] to [CHECKPOINT_PATH]
	`
)

type MigrateArgs struct {

	// Pod select options
	Namespace      string
	SourcePodName  string
	checkpointPath string
}

func NewPluginCmd() *cobra.Command {
	var Margs MigrateArgs
	cmd := &cobra.Command{
		Use:     "checkpoint [OPTIONS] POD_NAME CHECKPOINT_PATH",
		Short:   "checkpoint a Pod",
		Long:    longDesc,
		Example: example,
		Run: func(c *cobra.Command, args []string) {
			if err := Margs.Complete(c, args); err != nil {
				fmt.Println(err)
			}
			/*
				if err := opts.Validate(); err != nil {
					fmt.Println(err)
				}
				if err := opts.Run(); err != nil {
					fmt.Println(err)
				}
			*/
			if err := Margs.Run(); err != nil {
				fmt.Println(err)
			}
		},
	}
	cmd.Flags().StringVar(&Margs.Namespace, "namespace", "default",
		"default namespace is \"default\"")
	return cmd
}

func (a *MigrateArgs) Complete(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("error pod not specified")
	}
	if len(args) == 1 {
		return fmt.Errorf("destHost not specified")
	}

	a.SourcePodName = args[0]
	a.checkpointPath = args[1]
	return nil
}

func (a *MigrateArgs) Run() error {
	// Step1: Get sourcePod
	ctx := context.Background()
	config, _ := clientcmd.BuildConfigFromFlags("", "/home/dcn/fault-detection/docs/anisble-playbook/kubernetes-the-hard-way/admin.kubeconfig")
	clientset, _ := kubernetes.NewForConfig(config)
	pod, _ := clientset.CoreV1().Pods("default").Get(ctx, a.SourcePodName, metav1.GetOptions{})
	podsClient := clientset.CoreV1().Pods(corev1.NamespaceDefault)
	// Step2: Prepare the annotations
	action := "checkpoint"
	ann := pod.ObjectMeta.Annotations
	if ann == nil {
		ann = make(map[string]string)
	}
	ann["snapshotPolicy"] = action
	ann["snapshotPath"] = a.checkpointPath
	pod.ObjectMeta.Annotations = ann

	// Step3: Update the annotations
	if _, err := podsClient.Update(context.TODO(), pod, metav1.UpdateOptions{}); err != nil {
		return err
	}

	// Step4: Wait until checkpoint info are created

	container := pod.Spec.Containers[0].Name
	for {
		_, err := os.Stat(path.Join(a.checkpointPath, strings.Split(pod.Name, "-")[0], container, "descriptors.json"))
		if os.IsNotExist(err) {
			time.Sleep(1000 * time.Millisecond)
		} else {
			break
		}
	}
	// Step5: Stop checkpoint process
	ann["snapshotPolicy"] = ""
	ann["snapshotPath"] = ""
	pod.ObjectMeta.Annotations = ann
	if _, err := podsClient.Update(context.TODO(), pod, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func main() {
	cmd := NewPluginCmd()
	cmd.Execute()
}
