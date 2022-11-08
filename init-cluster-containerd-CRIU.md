## This document guides how to bootstrap Kubernetes cluster without Docker and enable CRIU intergration for Podmigration feature.
### Step0: 
- Ubuntu 18.04
- Do step 1-6 at all nodes.
- Do step 7 at controller node.
- Do step 8 at each worker nodes
### Step1: Install container runtime - Containerd
- Download containerd and unpackage:
```
$ sudo apt-get update
$ sudo apt-get install gcc

$ mkdir tmp
$ cd tmp/
$ sudo wget https://golang.org/dl/go1.15.5.linux-amd64.tar.gz
$ sudo tar -xzf go1.15.5.linux-amd64.tar.gz
$ sudo mv go /usr/local
$ sudo vi $HOME/.profile
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export GOBIN=$GOPATH/bin
export PATH=$GOROOT/bin:$GOBIN:$PATH
$ source $HOME/.profile
$ go version

$ sudo apt install make
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
$ git clone https://github.com/SSU-DCN/podmigration-operator.git
$ cd podmigration-operator
$ tar -vxf binaries.tar.bz2
$ cd custom-binaries/
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
$ sudo mv runc /usr/local/bin/
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

$ sudo -s
$ sudo echo '1' > /proc/sys/net/ipv4/ip_forward
$ exit
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
$ whereis kubelet
```

### Step5 : Replace kubelet with the custom kubelet.
The kubelet and kubeadm binaries can be found in this folder. (If you already download this file at Step1, just go to directory custom-binaries/)
```
$ git clone https://github.com/vutuong/kubernetes.git
$ git clone https://github.com/SSU-DCN/podmigration-operator.git
$ cd podmigration-operator
$ tar -vxf binaries.tar.bz2
$ cd custom-binaries
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
### Step7: You may need to change kubelet.config mode. (Or not)
```
$ sudo nano /var/lib/kubelet/config.yaml

authorization:
  #mode: Webhook
  mode: AlwaysAllow
```
### Step8: Install CRIU.
```
$ git clone https://github.com/SSU-DCN/podmigration-operator.git
$ cd podmigration-operator
$ tar -vxf criu-3.14.tar.bz2
$ cd criu-3.14
$ sudo apt-get install protobuf-c-compiler libprotobuf-c0-dev protobuf-compiler \
libprotobuf-dev:amd64 gcc build-essential bsdmainutils python git-core \
asciidoc make htop git curl supervisor cgroup-lite libapparmor-dev \
libseccomp-dev libprotobuf-dev libprotobuf-c0-dev protobuf-c-compiler \
protobuf-compiler python-protobuf libnl-3-dev libcap-dev libaio-dev \
apparmor libnet1-dev libnl-genl-3-dev libnl-route-3-dev libnfnetlink-dev pkg-config

$ make clean
$ make
$ sudo make install
$ criu check
$ criu check --all

$ mkdir /etc/criu
$ touch /etc/criu/runc.conf
$ nano /etc/criu/runc.conf
tcp-established
tcp-close
```
### Step9: Config NFS shared folder for every node in the cluster.
- Config NFS server at Master node
```
$ sudo apt-get update
$ sudo apt-get install nfs-kernel-server
$ sudo nano /etc/exports
/var/lib/kubelet/migration/  192.168.10.0/24(rw,sync,no_subtree_check)
Note: 192.168.10.0/24 is subnetmask of every node in our cluster
$ sudo exportfs -arvf
$ sudo systemctl start nfs-kernel-server
$ sudo systemctl enable nfs-kernel-server
$ sudo systemctl status nfs-kernel-server
$ sudo chmod 777 /var/lib/kubelet/migration
```
- Config NFS client at every worker nodes
```
$ sudo apt-get update
$ sudo apt-get install nfs-common
$ sudo nano /etc/fstab
192.168.10.13:/var/lib/kubelet/migration   /var/lib/kubelet/migration  nfs  defaults,_netdev 0 0
Note: 192.168.10.13 is the IP address of Nfs-server (master node) in this case.
$ sudo umount /var/lib/kubelet/migration
$ sudo mount -a
(If Mount Error occured, $ mount -t nfs -o nfsvers=3 <MASTER_NODE_IP>:/share /mnt)
$ sudo chmod 777 /var/lib/kubelet/migration
```
- Ref: https://github.com/vutuong/personal-notes/blob/master/configNFS.md
- You should remake the demo as the video in youtube after finish all the step above without```Step5-Approach 2```.
- Only if you don't need to use pre-build binaries of kubelet and kubeadm or you need to edit kubelet source code by your self and rebuild:
  ### Step5- Approach 2:  Download the custom source code and build.
  Download the custom source code and build at directory containerd/, the source code can be found as followings
  ```
  $ git clone https://github.com/vutuong/kubernetes.git
  ```
  You can find the binaries somewhere in the directories: ```kubernetes/kubernetes/_output/local/bin```
  - The link ref for setting up environments and build the custom binaries: https://www.youtube.com/watch?v=Q91iZywBzew&t=3509s&ab_channel=CNCF%5BCloudNativeComputingFoundation%5D





