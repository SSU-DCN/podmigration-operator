### Checkpoint and migrate commands in Kubernetes
In this repo, I write extensions for kubectl, which supports the commands of checkpoint and migrate a running pod in Kubernetes.
### How to build:
* To build checkpoint command:
```
$ cd checkpoint-command
$ go build -o kubectl-checkpoint
$ sudo cp kubectl-checkpoint /usr/local/bin
```
* To build migrate command:
```
$ cd migrate-command
$ go build -o kubectl-migrate
$ sudo cp kubectl-migrate /usr/local/bin
```
### How to use the checkpoint/migrate commands:
* To run the checkpoint command:
```
$ kubectl checkpoint --help
$ kubectl checkpoint [POD_NAME] [CHECKPOINT_PATH]
```
* Example of checkpoint a running pod:
```
dcn@dcn:~$ kubectl checkpoint --help

        checkpoint [POD_NAME] to [CHECKPOINT_PATH]

Usage:
  checkpoint [OPTIONS] POD_NAME CHECKPOINT_PATH [flags]

Examples:

        # Checkpoint a running Pod and save the checkpoint infomations to given path
        kubectl checkpoint [POD_NAME] [CHECKPOINT_PATH]
        kubectl checkpoint [POD_NAME] --namespace string [CHECKPOINT_PATH]


Flags:
  -h, --help               help for checkpoint
      --namespace string   default namespace is "default" (default "default")
```
```
dcn@dcn:~$ kubectl checkpoint simple /var/lib/kubelet/migration/xxx
```
* To run the migrate command:
```
dcn@dcn:~$ kubectl migrate --help
dcn@dcn:~$ kubectl migrate video worker1 
```
* Example of migrate a running pod:
```
dcn@dcn:~$ kubectl migrate --help

        migrate [POD_NAME] to [destHost]

Usage:
  migrate [OPTIONS] POD_NAME destHost [flags]

Examples:

        # Live-migrate a running Pod
        kubectl migrate [POD_NAME] [destHost]
        kubectl migrate [POD_NAME] --namespace string [destHost]


Flags:
  -h, --help               help for migrate
      --namespace string   default namespace is "default" (default "default")
```
```
dcn@dcn:~$ kubectl migrate video worker1
response Status: 200 OK
{
 "name": "video-migration-controller-70",
 "destHost": "",
 "replicas": 0,
 "selector": null,
 "action": "live-migration",
 "snapshotPath": "",
 "sourcePod": "video",
 "status": {
  "state": "",
  "currentRevision": "",
  "activePod": ""
 }
}
```
