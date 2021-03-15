package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/homedir"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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
	FromCluster   string
	ToCluster     string
}

type Podmigration struct {
	Name     string `json:"name"`
	DestHost string `json:"destHost"`
	Replicas int    `json:"replicas"`
	// Selector     *metav1.LabelSelector `json:"selector"`
	Action       string                 `json:"action"`
	SnapshotPath string                 `json:"snapshotPath"`
	SourcePod    string                 `json:"sourcePod"`
	Template     corev1.PodTemplateSpec `json:"template"`
	FromCluster  string                 `json:"fromCluster,omitempty"`
	ToCluster    string                 `json:"toCluster,omitempty"`
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
	cmd.Flags().StringVar(&Margs.DestHost, "desthost", "default",
		"default DestHost is \"null\"")
	cmd.Flags().StringVar(&Margs.SourcePodName, "sourcePodName", "default",
		"default sourcePodName is \"null\"")
	cmd.Flags().StringVar(&Margs.FromCluster, "fromCluster", "default",
		"default FromCluster is \"null\"")
	cmd.Flags().StringVar(&Margs.ToCluster, "toCluster", "default",
		"default ToCluster is \"null\"")
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

	// Step1: Define argument

	method := "POST"
	template := a.getSourcePodTemplate(a.SourcePodName, a.FromCluster)
	if template.ObjectMeta.Name == "" {
		fmt.Println("sourcePod: ", a.SourcePodName, " - does not exist in cluster ", a.FromCluster)
		return nil
	}
	fmt.Println(a.FromCluster)
	fmt.Println(a.ToCluster)
	if a.FromCluster != "" && a.ToCluster != "" {
		// Step1: Request to sourceCluster to checkpoint the pod
		sourceClusterURL := getClusterURL(a.FromCluster)
		podmigrationCheckpoint := &Podmigration{
			Name:      a.getCrdName(),
			DestHost:  a.DestHost,
			Replicas:  1,
			Action:    "checkpoint",
			SourcePod: a.SourcePodName,
			Template:  template,
			// FromCluster: a.FromCluster,
			// ToCluster:   a.ToCluster,
		}
		if err := requestAction(sourceClusterURL, method, podmigrationCheckpoint); err != nil {
			fmt.Println("Can'nt not reach sourceCluster", a.FromCluster)
			return err
		}

		// Step2: Request to destCluster to restore the pod
		destClusterURL := getClusterURL(a.ToCluster)
		podmigrationRestore := &Podmigration{
			Name:      a.getCrdName(),
			DestHost:  a.DestHost,
			Replicas:  1,
			Action:    "restore",
			SourcePod: a.SourcePodName,
			Template:  template,
			// FromCluster: a.FromCluster,
			// ToCluster:   a.ToCluster,
		}
		if err := requestAction(destClusterURL, method, podmigrationRestore); err != nil {
			fmt.Println("Can'nt not reach sourceCluster", a.ToCluster)
			return err
		}

	}
	// podmigration := &Podmigration{
	// 	Name:        a.getCrdName(),
	// 	DestHost:    a.DestHost,
	// 	Replicas:    1,
	// 	Action:      "live-migration",
	// 	SourcePod:   a.SourcePodName,
	// 	Template:    template,
	// 	FromCluster: a.FromCluster,
	// 	ToCluster:   a.ToCluster,
	// }
	// data := fmt.Sprintf(`{"name": "%s","action": "%s","sourcePod": "%s","destHost":"%s","template":"%s"}`, crdName, action, sourcePod, destHost, Template)
	return nil
}

func (a *MigrateArgs) getCrdName() string {
	s1 := rand.NewSource(time.Now().UnixNano())
	number := rand.New(s1)
	crdName := a.SourcePodName + "-migration-controller-" + strconv.Itoa(number.Intn(100))
	return crdName
}

func (a *MigrateArgs) getSourcePodTemplate(sourcePodName, fromCluster string) corev1.PodTemplateSpec {
	ctx := context.Background()
	// read the config file, so the plugin can talk to API-server
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", fromCluster), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	sourcePod, err := clientset.CoreV1().Pods("default").Get(ctx, sourcePodName, metav1.GetOptions{})
	if err != nil {
		return corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{},
			},
		}
	}
	pod := sourcePod.DeepCopy()
	container := pod.Spec.Containers[0]
	automountServiceAccountToken := false
	template := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:   pod.ObjectMeta.Name,
			Labels: pod.ObjectMeta.Labels,
		},
		Spec: corev1.PodSpec{
			ServiceAccountName:           "default",
			AutomountServiceAccountToken: &automountServiceAccountToken,
			Containers: []corev1.Container{
				{
					Name:         container.Name,
					Image:        container.Image,
					Ports:        container.Ports,
					VolumeMounts: container.VolumeMounts,
				},
			},
			Volumes: pod.Spec.Volumes,
		},
	}
	return template
}
func requestAction(url, method string, podmigration *Podmigration) error {
	rawData, _ := json.Marshal(podmigration)
	data := fmt.Sprintf(`%s`, rawData)
	payload := strings.NewReader(data)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		fmt.Println(err)
	}
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

func getClusterURL(cluster string) string {
	if cluster == "cluster1" {
		return "http://192.168.10.99:5000/Podmigrations"
	} else if cluster == "cluster2" {
		return "http://192.168.10.42:5000/Podmigrations"
	} else {
		return ""
	}
}
func main() {
	cmd := NewPluginCmd()
	cmd.Execute()
}
