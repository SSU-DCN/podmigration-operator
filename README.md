# podmigration-operator
## The document to init K8s cluster, which enables Podmigration, can be found at: 
- https://github.com/SSU-DCN/podmigration-operator/blob/main/init-cluster-containerd-CRIU.md

## How to run:
(Run this command at directory podmigration-operator/)
* To run Podmigration operator, which includes CRD and a custom controller:
```
$ sudo snap install kustomize
$ sudo apt-get install gcc
$ make manifests
$ make install
$ make run
```
* To run api-server, which enables ```kubectl migrate``` command and GUI: (at podmigration-operator/ directory)
```
$ go run ./api-server/cmd/main.go
```
* To install ```kubectl migrate/checkpoint``` command, follow the guide at https://github.com/SSU-DCN/podmigration-operator/tree/main/kubectl-plugin
* To run GUI:
```
$ cd podmigration-operator/gui
$ npm install
$ npm run serve
```
### Demo video:
1. Migrate video streaming pod from node to node in single cluster:
 -  https://www.youtube.com/watch?v=M4Ik7aUKhas&t=1s&ab_channel=Xu%C3%A2nT%C6%B0%E1%BB%9DngV%C5%A9
2. Migrate video streaming pod from cluster to cluster:
 -  https://www.youtube.com/watch?v=Bpdlgu0XZqo
 - https://drive.google.com/file/d/1AeyJZTRJcayBelvXf-CZwFapoquBpns1/view?usp=sharing

## Test live-migrate pod:
* Run/check video-stream application:
```
$ cd podmigration-operator/config/samples/migration-example
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
- Note: As default, K8S doesn't have ```kubectl migrate``` and ```kubectl checkpoint``` command. To use this extended kubectl plugin please check the guide at https://github.com/SSU-DCN/podmigration-operator/tree/main/kubectl-plugin
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

