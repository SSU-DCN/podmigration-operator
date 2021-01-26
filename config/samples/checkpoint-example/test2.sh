# Test1: with Redmine
kubectl apply -f redmine-restore.yaml
time (while ! curl http://192.168.10.13:31764/ > /dev/null 2>&1; do : ; done )
kubectl delete -f redmine-restore.yaml

# Test1: with Redmine
# kubectl apply -f ruby-restore.yaml
# time (while ! curl http://192.168.10.13:30087/ > /dev/null 2>&1; do : ; done )
# kubectl delete -f ruby-restore.yaml
