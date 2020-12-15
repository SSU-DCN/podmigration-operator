package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	defaultConfigPath = "/home/dcn/fault-detection/docs/anisble-playbook/kubernetes-the-hard-way/admin.kubeconfig"
	defaultNameSpace  = "default"
	pingPodImage      = "ianneub/network-tools:latest"
)

type delayvalue struct {
	value    []float64
	from, to string
}

type latencyMonitor struct {
	// dlMatrix []*delayvalue
	method string
	client *kubernetes.Clientset
	config *rest.Config
}

// each pingPod have a nodeParent and ip
type pingPod struct {
	podName    string
	nodeParent string
	podIP      string
}

// get all worker node that is Ready
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

// prepare Ping pod template for deploying
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

// deploy Ping Pod in the nodes
func (lm *latencyMonitor) deployPingPod(expectedNode string) error {
	podName := "pingpod-" + expectedNode
	pod := lm.getPodTemplate(podName, expectedNode)
	// now create the pod in kubernetes cluster using the clientset
	pod, err := lm.client.CoreV1().Pods(pod.Namespace).Create(pod)
	if err != nil {
		// panic(err)
		return err
	}
	fmt.Println("Pod created successfully...")
	return nil
}

// check Ping Pod is Running or not in a particular node
func (lm *latencyMonitor) checkPingPod(expectedNode string) bool {
	pods, err := lm.client.CoreV1().Pods("").List(metav1.ListOptions{
		LabelSelector: "app=ping",
		FieldSelector: "spec.nodeName=" + expectedNode,
	})
	if err != nil {
		log.Printf("Could not get PingPod info: %+v", err)
		return false
	}
	if pods.Items[0].Status.Phase == "Running" {
		log.Printf("Pod check ok with phase: %+v", pods.Items[0].Status.Phase)
		// log.Printf("Could not get PingPod info:%+v", pod.Status)
		return true
	} else {
		log.Printf("Pod check fail with phase: %+v", pods.Items[0].Status.Phase)
		// log.Printf("Could not get PingPod info:%+v", pod.Status)
		return false
	}
}

// get IP of PingPod of a particular node
func (lm *latencyMonitor) getPodIP(expectedNode string) string {
	pod, _ := lm.client.CoreV1().Pods("default").List(metav1.ListOptions{
		LabelSelector: "app=ping",
		FieldSelector: "spec.nodeName=" + expectedNode,
	})
	log.Printf("Pod : %+v", pod.Items[0].Status.PodIP)
	return pod.Items[0].Status.PodIP
}

func (lm *latencyMonitor) measureRTT(sourcePod, destPod *pingPod, config *rest.Config) (*string, error) {
	command := []string{
		"/bin/sh",
		"-c",
		fmt.Sprintf(`ping -c 5 %s`, destPod.podIP),
	}
	req := lm.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(sourcePod.podName).
		Namespace("default").
		SubResource("exec")
	req.VersionedParams(&core.PodExecOptions{
		Container: "ping",
		Command:   command,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	log.Printf("Request URL: %s", req.URL().String())

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		log.Printf("Failed to exec:%v", err)
		return nil, err
	}
	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		log.Printf("Faile to get result:%v", err)
		return nil, err
	}
	// return stdout.Bytes(), stderr.Bytes(), nil
	// regex to find rtt
	str := strings.Split(string(stdout.Bytes()), "\n")
	re := regexp.MustCompile(`(?m)(\d.\d+)/(\d.\d+)/(\d.\d+)/(\d.\d+)`)
	match := re.FindAllString(str[len(str)-2], -1)
	rttLine := strings.Split(string(match[0]), "/")
	// minRtt := rttLine)[0]
	avgRtt := rttLine[1]
	// maxRtt := rttLine[2]
	// return strconv.ParseFloat(avgRtt, 32), nil
	return &avgRtt, nil
}

func (lm *latencyMonitor) fileWrite() {
	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(number int) {
			f, err := os.OpenFile("/home/dcn/podmigration-operator/latency-monitor/test.file.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				log.Printf("err")
			}
			for j := 0; j < 10; j++ {
				r := strings.NewReader(fmt.Sprintf("goroutine: %d, loop: %d\n", number, j))
				_, err = io.Copy(f, r)
				if err != nil {
					log.Printf("err")
				}
			}
			err = f.Close()
			if err != nil {
				log.Printf("err")
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func main() {
	pingPodList := []*pingPod{}
	clientset, config, err := initClientset(defaultConfigPath)
	if err != nil {
		panic(err)
	}

	// init latencyMonitor object
	monitor := latencyMonitor{method: "Ping", client: clientset, config: config}
	size, _ := monitor.getWorkerNode()

	for _, node := range size {
		isPingPodRunning := monitor.checkPingPod(node)
		if isPingPodRunning == false {
			monitor.deployPingPod(node)
		}
		// monitor.getPodIP(node)
		pingPodOfNode := pingPod{podName: "pingpod-" + node, nodeParent: node, podIP: monitor.getPodIP(node)}
		pingPodList = append(pingPodList, &pingPodOfNode)
	}
	log.Printf("The number of nodes in cluster:%+v", len(size))
	log.Printf("The nodes in our clusters:%+v", pingPodList[0].podIP)

	// sourcePod := pingPod{podName: "pingpod-" + "worker1", nodeParent: "worker1", podIP: "10.22.0.21"}
	// destPod := pingPod{podName: "pingpod-" + "worker1", nodeParent: "worker1", podIP: "10.22.0.21"}

	// rtt, err := monitor.measureRTT(&sourcePod, &destPod, monitor.config)
	// if err != nil {
	// 	log.Printf("Failed to exec:%v", err)
	// 	// return map[string]int{}, err
	// }
	// log.Printf("out:%s", *rtt)
	// monitor.fileWrite("fff")

}

func initClientset(configPath string) (*kubernetes.Clientset, *rest.Config, error) {
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
		// panic(err)
		os.Exit(1)
	}
	return clientset, config, nil
}
