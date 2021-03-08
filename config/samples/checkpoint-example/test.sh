# # test1 with Redmine
kubectl apply -f redmine.yaml
time (while ! curl http://192.168.10.13:31764/ > /dev/null 2>&1; do : ; done )
kubectl checkpoint tuongvx /var/lib/kubelet/migration/ooo
kubectl delete -f redmine.yaml

# # test with Web Rail application
# kubectl apply -f ruby.yaml
# time (while ! curl http://192.168.10.13:30087/ > /dev/null 2>&1; do : ; done ) > eval.txt
# kubectl checkpoint tuongvx /var/lib/kubelet/migration/fff
# kubectl delete -f ruby.yaml

#kubectl apply -f video.yaml
#time (while ! curl http://192.168.10.13:31764/ > /dev/null 2>&1; do : ; done )
#kubectl checkpoint tuongvx /var/lib/kubelet/migration/ooo
#kubectl delete -f redmine.yaml
