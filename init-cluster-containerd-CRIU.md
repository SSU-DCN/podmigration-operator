## This document guides how to bootstrap Kubernetes cluster without Docker and enable CRIU intergration for Podmigration feature.
### Step0: 
- Ubuntu 18.04
- Do step 1-5 in every node.
- Do step 6 in controller node.
### Step1: Install container runtime - Containerd
- Download containerd and unpackage:
```
$ sudo apt-get update
$ wget https://github.com/containerd/containerd/releases/download/v1.3.6/containerd-1.3.6-linux-amd64.tar.gz
$ mkdir containerd
$ tar -xvf containerd-1.3.6-linux-amd64.tar.gz -C containerd
$ sudo mv containerd/bin/* /bin/
```
- Replace the containerd-cri with interface extentions supporting CRIU, which is need for running podmigration. There are two options:
1. Build from source:
```
$ git clone https://github.com/vutuong/containerd-cri.git
$ cd containerd-cri/
$ sudo snap install go --classic
$ go version
$ go get github.com/containerd/cri/cmd/containerd
$ make
$ sudo make install
$ cd _output/
$ sudo mv containerd /bin/
```
2. Download binaries:
```
$ cd containerd/
$ wget https://k8s-pod-migration.obs.eu-de.otc.t-systems.com/v2/containerd
$ chmod +x containerd
$ sudo mv containerd /bin/
```
- Configure containerd and create the containerd configuration file
```
$ sudo mkdir /etc/containerd
$ sudo nano /etc/containerd/config.toml

[plugins]
  [plugins.cri.containerd]
    snapshotter = "overlayfs"
    [plugins.cri.containerd.default_runtime]
      runtime_type = "io.containerd.runtime.v1.linux"
      runtime_engine = "/usr/local/bin/runc"
      runtime_root = ""
```
- Install newest version of runc, at this time, I'm using v1.0.0-rc92 (the podmigration)
```
$ wget https://github.com/opencontainers/runc/releases/download/v1.0.0-rc92/runc.amd64
$ whereis runc
$ sudo mv runc.amd64 runc
$ chmod +x runc
$ sudo mv runc usr/local/bin/
```
- Configure containerd and create the containerd.service systemd unit file
```
$ sudo nano /etc/systemd/system/containerd.service

[Unit]
Description=containerd container runtime
Documentation=https://containerd.io
After=network.target

[Service]
ExecStartPre=/sbin/modprobe overlay
ExecStart=/bin/containerd
Restart=always
RestartSec=5
Delegate=yes
KillMode=process
OOMScoreAdjust=-999
LimitNOFILE=1048576
LimitNPROC=infinity
LimitCORE=infinity

[Install]
WantedBy=multi-user.target
```
- Reload containerd service
```
$ sudo systemctl daemon-reload
$ sudo systemctl restart containerd
$ sudo systemctl status containerd
```
### Step2: Solve a few problems introduced with containerd
```
$ sudo nano /etc/sysctl.conf
...
net.bridge.bridge-nf-call-iptables = 1

$ sudo echo '1' > /proc/sys/net/ipv4/ip_forward
$ sudo sysctl --system
$ sudo modprobe overlay
$ sudo modprobe br_netfilter
```
### Step3: You'll need to map all of your nodes in ```/etc/hosts```
### Step4: Install kubernetes components
```
$ curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add
$ sudo apt-add-repository "deb http://apt.kubernetes.io/ kubernetes-xenial main"
$ sudo apt-get install kubeadm=1.19.0-00 kubelet=1.19.0-00 kubectl=1.19.0-00 -y
$ whereis kubeadm
$ whereis kubele
$ whereis kubelet
```
### Step5: Replace kubelet with the custom kubelet. There are two options:
1. Build from source, download the custom source code and build, the source code can be found as followings
```
$ git clone https://github.com/vutuong/kubernetes.git
```
- The link ref for setting up environments: https://www.youtube.com/watch?v=Q91iZywBzew&t=3509s&ab_channel=CNCF%5BCloudNativeComputingFoundation%5D
After all, run ```make```, and you can find the binaries somewhere in the directories: ```kubernetes/kubernetes/_output/local/bin```
2. The kubelet and kubeadm binaries can be found in this folder.
```
$ git clone https://github.com/SSU-DCN/podmigration-operator.git
$ cd podmigration-operator/custom-binaries
```
- After downloading/building the kubelet and kubeadm binaries, replace it with existing default K8s-kubelet binaries:
```
$ chmod +x kubeadm kubelet
$ sudo mv kubeadm kubelet /usr/bin/
$ sudo systemctl daemon-reload
$ sudo systemctl restart kubelet
$ sudo systemctl status kubelet
```
### Step6: Init k8s-cluster. 
- In the master node run following command:
```
$ sudo kubeadm init --pod-network-cidr=10.244.0.0/16
```
- Use the output log command to join workernode
- To run the kubectl commands:
```
$ kubectl get nodes
$ mkdir -p $HOME/.kube
$ sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
$ sudo chown $(id -u):$(id -g) $HOME/.kube/config
```
- Apply flannel CNI add-on network:
```
$ kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml
$ kubectl get pods -n kube-system
```
- Check custom kubelet is running:
```
$ journalctl -fu kubelet

Feb 04 05:55:16 tuong-worker1 kubelet[26650]: I0204 05:55:16.979326   26650 kuberuntime_manager.go:841] Should we migrate?Runningfalse
Feb 04 05:55:21 tuong-worker1 kubelet[26650]: I0204 05:55:21.979185   26650 kuberuntime_manager.go:841] Should we migrate?Runningfalse
Feb 04 05:55:25 tuong-worker1 kubelet[26650]: I0204 05:55:25.979207   26650 kuberuntime_manager.go:841] Should we migrate?Runningfalse

```


