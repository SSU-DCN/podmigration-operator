# podmigration-operator

## Kubebuilder init command
```
kubebuilder init --domain dcn.ssu.ac.kr
```
```
kubebuilder create api --group podmig --version v1 --kind Podmigration
```

## How to run:
* To run operator:
```
make run
```
* To run api-server :
```
go run ./api-server/cmd/main.go
```
## Test live-migrate pod:
* Run/check video-stream application:
```
$ cd podmigration-operator/config/samples
$ kubectl apply -f 2.yaml
$ kubectl get pods
```
* Live-migrate video-stream application via api-server:
```
$ curl --request POST 'localhost:5000/Podmigrations' --header 'Content-Type: application/json' --data-raw '{"name":"test1", "replicas":1, "action":"live-migration", "sourcePod":"video", "destHost":"worker1"}'
$ curl --request GET 'localhost:5000/Podmigrations'
```
* Live-migrate video-stream application via kubectl:
```
$ kubectl apply -f test2.yaml
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

