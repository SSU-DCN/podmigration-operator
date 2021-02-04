# podmigration-operator
## The documents to init K8s cluster,which enable Podmigration, can be found at: 
- https://github.com/SSU-DCN/podmigration-operator/blob/main/init-cluster-containerd-CRIU.md
## Kubebuilder init command
```
kubebuilder init --domain dcn.ssu.ac.kr
```
```
kubebuilder create api --group podmig --version v1 --kind Podmigration
```

## How to run:
* To run Podmigration operator, which include CRD and custom controller:
```
$ make run
```
* To run api-server, which enable ```kubectl migrate``` command and GUI:
```
$ go run ./api-server/cmd/main.go
```
* To run GUI:
```
$ cd podmigration-operator/gui
$ npm run serve
```
## Test live-migrate pod:
* Run/check video-stream application:
```
$ cd podmigration-operator/config/samples
$ kubectl apply -f 2.yaml
$ kubectl get pods
```
#### There are three options to live-migrate a running Pod as following:
1. Live-migrate video-stream application via api-server:
```
$ curl --request POST 'localhost:5000/Podmigrations' --header 'Content-Type: application/json' --data-raw '{"name":"test1", "replicas":1, "action":"live-migration", "sourcePod":"video", "destHost":"worker1"}'
$ curl --request GET 'localhost:5000/Podmigrations'
```
2. Live-migrate video-stream application via kubectl apply:
```
$ kubectl apply -f test2.yaml
```
3. Live-migrate video-stream application via kubectl migrate command:
```
$ kubectl migrate video worker1
```
* To delete:
```
$ kubectl delete podmigration test2
$ kubectl delete -f test2.yaml
```
## Note
This operator is controller of Kuberntes Pod migration for Kubernetes. It needs several changes to work such as: kubelet, container-runtime-cri (containerd-cri). The modified vesions of Kuberntes and containerd-cri beside this operator can be found in the following repos:

* https://github.com/vutuong/kubernetes


* https://github.com/vutuong/containerd-cri

## References
* https://github.com/kubernetes/kubernetes/issues/3949

## Workflow
![alt text](https://github.com/SSU-DCN/podmigration-operator/blob/main/podmigration.jpg?raw=true)

