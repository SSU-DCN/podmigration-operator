package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	// "k8s.io/client-go/kubernetes"
	// "k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/* To use this kubectl-plugin
 		$ go build -o kubectl-migrate
		$ sudo cp kubectl-migrate /usr/local/bin
*/

const (
	example = `
	# Live-migrate a running Pod
	kubectl migrate [POD_NAME] [destHost]
	kubectl migrate [POD_NAME] --namespace string [destHost]
	`
	longDesc = `
	migrate [POD_NAME] to [destHost]
	`
)

type PodmigrationEndpoint struct {
	client client.Client
}

type MigrateArgs struct {

	// Pod select options
	Namespace     string
	SourcePodName string
	DestHost      string
}

type Podmigration struct {
	Name     string `json:"name"`
	DestHost string `json:"destHost"`
	Replicas int    `json:"replicas"`
	// Selector     *metav1.LabelSelector `json:"selector"`
	Action       string `json:"action"`
	SnapshotPath string `json:"snapshotPath"`
	SourcePod    string `json:"sourcePod"`
	// Template corev1.PodTemplateSpec `json:"template"`
}

func NewPluginCmd() *cobra.Command {
	var Margs MigrateArgs
	cmd := &cobra.Command{
		Use:     "migrate [OPTIONS] POD_NAME destHost",
		Short:   "migrate a Pod",
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
	a.DestHost = args[1]
	return nil
}

func (a *MigrateArgs) Run() error {
	// ctx := context.Background()
	//read the config file, so the plugin can talk to API-server
	// config, _ := clientcmd.BuildConfigFromFlags("", "/home/dcn/fault-detection/docs/anisble-playbook/kubernetes-the-hard-way/admin.kubeconfig")
	// clientset, _ := kubernetes.NewForConfig(config)
	// pod, _ := clientset.CoreV1().Pods("default").Get(ctx, "pingpod-worker1", metav1.GetOptions{})

	// Step1: Define argument
	url := "http://localhost:5000/Podmigrations"
	method := "POST"
	crdName := a.getCrdName()
	action := "live-migration"
	sourcePod := a.SourcePodName
	destHost := a.DestHost
	data := fmt.Sprintf(`{"name": "%s","action": "%s","sourcePod": "%s","destHost":"%s"}`, crdName, action, sourcePod, destHost)
	payload := strings.NewReader(data)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("contentType", "application/json")
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	body, err := ioutil.ReadAll(res.Body)
	fmt.Println("response Status:", res.Status)
	fmt.Println(string(body))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func (a *MigrateArgs) getCrdName() string {
	s1 := rand.NewSource(time.Now().UnixNano())
	number := rand.New(s1)
	crdName := a.SourcePodName + "-migration-controller-" + strconv.Itoa(number.Intn(100))
	return crdName
}

func main() {
	cmd := NewPluginCmd()
	cmd.Execute()
}
