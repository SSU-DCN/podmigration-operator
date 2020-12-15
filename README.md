# podmigration-operator

## Kubebuilder init command
```
kubebuilder init --domain dcn.ssu.ac.kr
```
```
kubebuilder create api --group podmig --version v1 --kind Podmigration
```
## Note
This operator is controller of Kuberntes Pod migration for Kubernetes. It needs several changes to work such as: kubelet, container-runtime-cri (containerd-cri). The modified vesions of Kuberntes and containerd-cri beside this operator can be found in the following repos:

* https://github.com/vutuong/kubernetes


* https://github.com/vutuong/containerd-cri

## References
* https://github.com/kubernetes/kubernetes/issues/3949

## Workflow
![alt text](https://github.com/SSU-DCN/podmigration-operator/blob/main/podmigration.jpg?raw=true)

