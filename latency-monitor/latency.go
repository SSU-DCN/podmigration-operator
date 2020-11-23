package main

import (
	"fmt"
	"log"
	"os"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultConfigPath = "/home/dcn/fault-detection/docs/anisble-playbook/kubernetes-the-hard-way/admin.kubeconfig"
	defaultNameSpace  = "default"
	pingPodImage      = "ianneub/network-tools:latest"
)

type delayvalue struct {
	value       []float64
	row, column string
}

type latencyMonitor struct {
	// dlMatrix []*delayvalue
	method string
	client *kubernetes.Clientset
}

func (lm *latencyMonitor) getWorkerNode() ([]string, error) {
	// ctx := context.Context()
	log.Printf("getWorkerNode function")
	nodes := []string{}

	nodeList, err := lm.client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		log.Printf("Unable to list cluster nodes")
		return nil, err
	}
	for _, node := range nodeList.Items {
		fmt.Println(node.Status.Conditions[3].Type)

		// check node Ready or not
		if node.Status.Conditions[len(node.Status.Conditions)-1].Type != "Ready" {
			log.Printf("Node not Ready:%+v", node.Name)
			return nil, nil
		}
		nodes = append(nodes, node.Name)
	}
	return nodes, nil
}

func (lm *latencyMonitor) getPodTemplate(podName, expectedNode string) *core.Pod {
	return &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: defaultNameSpace,
			Labels: map[string]string{
				"app": "ping",
			},
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:            "ping",
					Image:           pingPodImage,
					ImagePullPolicy: core.PullIfNotPresent,
					Command: []string{
						"/sbin/init",
					},
				},
			},
			NodeName: expectedNode,
		},
	}
}

func (lm *latencyMonitor) deployPingPod(expectedNode string) {
	podName := "pingpod-" + expectedNode
	pod := lm.getPodTemplate(podName, expectedNode)
	// now create the pod in kubernetes cluster using the clientset
	pod, err := lm.client.CoreV1().Pods(pod.Namespace).Create(pod)
	if err != nil {
		panic(err)
	}
	fmt.Println("Pod created successfully...")
}

func main() {

	clientset, err := initClientset(defaultConfigPath)
	if err != nil {
		panic(err)
	}

	// init latencyMonitor object
	monitor := latencyMonitor{method: "Ping", client: clientset}
	size, _ := monitor.getWorkerNode()

	for _, node := range size {
		monitor.deployPingPod(node)
	}
	log.Printf("The number of nodes in cluster:%+v", len(size))
	log.Printf("The nodes in our clusters:%+v", size)
}

func initClientset(configPath string) (*kubernetes.Clientset, error) {
	// var kubeconfig *string
	// if home := homedir.HomeDir(); home != "" {
	// 	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	// } else {
	// 	kubeconfig = flag.String("kubeconfig", "", configPath)
	// }
	// flag.Parse()

	// config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	// if err != nil {
	// 	panic(err)
	// }
	// config, _ := clientcmd.BuildConfigFromFlags("", configPath)
	// clientset, err := kubernetes.NewForConfig(config)
	config, _ := clientcmd.BuildConfigFromFlags("", configPath)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
	return clientset, nil
}
