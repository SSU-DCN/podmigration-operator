/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	podmigv1 "github.com/SSU-DCN/podmigration-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"

	corev1 "k8s.io/api/core/v1"
	// corev1 "github.com/vutuong/kubernetes/tree/feature/pod-migration/staging/src/k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// podOwnerKey = ".metadata.controller"
	podOwnerKey = "migratingPod"
	// migratingPodFinalizer = "podmig.schrej.net/Migrate"
)

// PodmigrationReconciler reconciles a Podmigration object
type PodmigrationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=podmig.dcn.ssu.ac.kr,resources=podmigrations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=podmig.dcn.ssu.ac.kr,resources=podmigrations/status,verbs=get;update;patch

func (r *PodmigrationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("podmigration", req.NamespacedName)

	// your logic here
	// Load the podMigration resource object, if there is no Object, return directly
	var migratingPod podmigv1.Podmigration
	if err := r.Get(ctx, req.NamespacedName, &migratingPod); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("", "print test", migratingPod.Spec)

	template := &migratingPod.Spec.Template
	desiredLabels := getPodsLabelSet(template)
	desiredLabels["migratingPod"] = migratingPod.Name
	annotations := getPodsAnnotationSet(&migratingPod)

	// Then list all pods controlled by the Podmigration resource object
	var childPods corev1.PodList
	if err := r.List(ctx, &childPods, client.InNamespace(req.Namespace), client.MatchingLabels(desiredLabels)); err != nil {
		log.Error(err, "unable to list child pods")
		return ctrl.Result{}, err
	}

	pod, err := r.desiredPod(migratingPod, &migratingPod, req.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}

	depl, err := r.desiredDeployment(migratingPod, &migratingPod, req.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}

	log.Info("", "annotations ", annotations["snapshotPath"])
	log.Info("", "number of existing pod ", len(childPods.Items))
	log.Info("", "desired pod ", pod)
	log.Info("", "number of desired pod ", migratingPod.Spec.Replicas)

	count, _, _ := r.getActualRunningPod(&childPods)
	log.Info("", "number of actual running pod ", count)

	if annotations["snapshotPolicy"] == "live-migration" && annotations["sourcePod"] != "" {
		// list source pods for live-migration
		sourcePod, err := r.checkPodExist(ctx, annotations["sourcePod"], req.Namespace)
		if err != nil || sourcePod == nil {
			log.Error(err, "sourcePod not exist", "pod", annotations["sourcePod"])
			return ctrl.Result{}, err
		}
		log.Info("", "sourcePod ok ", sourcePod)
		log.Info("", "sourcePod status ", sourcePod.Status.Phase)

		// Step1: checkpoint sourcePod
		copySourcePod := sourcePod.DeepCopy()
		if err := r.checkpointPod(ctx, copySourcePod); err != nil {
			log.Error(err, "unable to checkpoint", "pod", sourcePod)
			return ctrl.Result{}, err
		}
		log.Info("", "Live-migration", "Step1 - checkpoint source Pod - completed")
		// for container := range copySourcePod.Spec.Containers {
		// 	fmt.Println(copySourcePod.Spec.Containers[container].Name)
		// 	log.Info("", "container of pod", copySourcePod.Spec.Containers[container].Name)
		// }

		// Step2: wait until checkpoint info are created
		container := copySourcePod.Spec.Containers[0].Name
		checkpointPath := path.Join("/var/lib/kubelet/migration/kkk", strings.Split(copySourcePod.Name, "-")[0])
		log.Info("", "live-migration pod", container)
		for {
			_, err := os.Stat(path.Join(checkpointPath, container, "descriptors.json"))
			if os.IsNotExist(err) {
				time.Sleep(1000 * time.Millisecond)
			} else {
				break
			}
		}
		log.Info("", "Live-migration", "checkpointPath"+checkpointPath)
		log.Info("", "Live-migration", "Step 2 - Wait until checkpoint info are created - completed")
		// time.Sleep(10)
		// Step3: restore destPod from sourcePod checkpoted info
		newPod, err := r.restorePod(ctx, pod, annotations["sourcePod"], checkpointPath)
		if err != nil {
			log.Error(err, "unable to restore", "pod", sourcePod)
			return ctrl.Result{}, err
		}
		log.Info("", "Live-migration", "Step 3 - Restore destPod from sourcePod's checkpointed info - completed")
		// time.Sleep(5)
		// Step4: Clean checkpointpod process and checkpointPath
		if err := r.removeCheckpointPod(ctx, copySourcePod, "/var/lib/kubelet/migration/kkk", newPod.Name, req.Namespace); err != nil {
			log.Error(err, "unable to remove checkpoint", "pod", sourcePod)
			return ctrl.Result{}, err
		}
		log.Info("", "Live-migration", "Step 4 - Clean checkpointPod process and checkpointPath completed")

		// step5: Delete source Pod
		if err := r.deletePod(ctx, sourcePod); err != nil {
			log.Error(err, "unable to delete", "source pod", sourcePod)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	if count == 0 {
		_, err := os.Stat(annotations["snapshotPath"])
		if annotations["snapshotPolicy"] != "restore" && os.IsNotExist(err) {
			pod.ObjectMeta.Annotations["snapshotPolicy"] = ""
			pod.ObjectMeta.Annotations["snapshotPath"] = ""
		}
		if err := r.createMultiPod(ctx, migratingPod.Spec.Replicas, depl); err != nil {
			log.Error(err, "unable to create Pod for restore", "pod", pod)
			return ctrl.Result{}, err
		}
	} else if count != 0 && count != migratingPod.Spec.Replicas {
		_, err := os.Stat(annotations["snapshotPath"])
		if annotations["snapshotPolicy"] != "restore" && os.IsNotExist(err) {
			pod.ObjectMeta.Annotations["snapshotPolicy"] = ""
			pod.ObjectMeta.Annotations["snapshotPath"] = ""
		}
		if err := r.updateMultiPod(ctx, migratingPod.Spec.Replicas-count, depl); err != nil {
			log.Error(err, "unable to create Pod for restore", "pod", pod)
			return ctrl.Result{}, err
		}
	} else {
		_, err := os.Stat(annotations["snapshotPath"])
		if annotations["snapshotPolicy"] == "checkpoint" && !os.IsNotExist(err) {
			curPod := &childPods.Items[0]
			newPod := curPod.DeepCopy()
			ann := newPod.ObjectMeta.Annotations
			if ann == nil {
				ann = make(map[string]string)
			}
			ann["snapshotPolicy"] = "checkpoint"
			ann["snapshotPath"] = annotations["snapshotPath"]
			newPod.ObjectMeta.Annotations = ann
			if err := r.Update(ctx, newPod); err != nil {
				log.Error(err, "unable to patch annotations", "pod", pod)
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
}

func (r *PodmigrationReconciler) getActualRunningPod(childPods *corev1.PodList) (int, corev1.PodList, corev1.PodList) {
	// if a pod is deleted, remove it from Actual running pod list
	count := 0
	var actualRunningPod, isDeletingPod corev1.PodList
	for _, pod := range childPods.Items {
		if !pod.DeletionTimestamp.IsZero() {
			isDeletingPod.Items = append(isDeletingPod.Items, pod)
		} else {
			actualRunningPod.Items = append(actualRunningPod.Items, pod)
			count++
		}
	}
	return count, actualRunningPod, isDeletingPod
}

func (r *PodmigrationReconciler) createMultiPod(ctx context.Context, replicas int, depl *appsv1.Deployment) error {
	if err := r.Create(ctx, depl); err != nil {
		return err
	}
	return nil
}

func (r *PodmigrationReconciler) updateMultiPod(ctx context.Context, replicas int, depl *appsv1.Deployment) error {
	if err := r.Update(ctx, depl); err != nil {
		return err
	}
	return nil
}

func (r *PodmigrationReconciler) deletePod(ctx context.Context, pod *corev1.Pod) error {
	if err := r.Delete(ctx, pod); err != nil {
		return err
	}
	return nil
}

func (r *PodmigrationReconciler) checkpointPod(ctx context.Context, pod *corev1.Pod) error {
	snapshotPolicy := "checkpoint"
	snapshotPath := "/var/lib/kubelet/migration/kkk"
	if err := r.updateAnnotations(ctx, pod, snapshotPolicy, snapshotPath); err != nil {
		return err
	}
	return nil
}

func (r *PodmigrationReconciler) restorePod(ctx context.Context, pod *corev1.Pod, sourcePod, checkpointPath string) (*corev1.Pod, error) {
	// targetPod := pod.DeepCopy()
	// targetPod.Finalizers = append(targetPod.Finalizers, migratingPodFinalizer)
	s1 := rand.NewSource(time.Now().UnixNano())
	number := rand.New(s1)
	sourcePod = strings.Split(sourcePod, "-migration-")[0]
	pod.Name = sourcePod + "-migration-" + strconv.Itoa(number.Intn(100))
	// pod.Spec.ClonePod = sourcePod
	pod.ObjectMeta.Annotations["snapshotPolicy"] = "restore"
	pod.ObjectMeta.Annotations["snapshotPath"] = checkpointPath
	if err := r.Create(ctx, pod); err != nil {
		return nil, err
	}
	return pod, nil
}

func (r *PodmigrationReconciler) removeCheckpointPod(ctx context.Context, pod *corev1.Pod, snapshotPathCurrent, newPodName, namespace string) error {
	for {
		ok, _ := r.checkPodExist(ctx, newPodName, namespace)
		if ok != nil {
			break
		}
	}
	snapshotPolicyUpdate := ""
	snapshotPathUpdate := ""
	if err := r.updateAnnotations(ctx, pod, snapshotPolicyUpdate, snapshotPathUpdate); err != nil {
		return err
	}
	os.Chmod(snapshotPathCurrent, 0777)
	if _, err := exec.Command("sudo", "rm", "-rf", snapshotPathCurrent).Output(); err != nil {
		return err
	}
	return nil
}

func (r *PodmigrationReconciler) updateAnnotations(ctx context.Context, pod *corev1.Pod, snapshotPolicy, snapshotPath string) error {
	ann := pod.ObjectMeta.Annotations
	if ann == nil {
		ann = make(map[string]string)
	}
	ann["snapshotPolicy"] = snapshotPolicy
	ann["snapshotPath"] = snapshotPath
	pod.ObjectMeta.Annotations = ann
	if err := r.Update(ctx, pod); err != nil {
		return err
	}
	return nil
}

func (r *PodmigrationReconciler) checkPodExist(ctx context.Context, name, namespace string) (*corev1.Pod, error) {
	var childPods corev1.PodList
	if err := r.List(ctx, &childPods, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	if len(childPods.Items) > 0 {
		for _, pod := range childPods.Items {
			if pod.Name == name && pod.Status.Phase == "Running" {
				return &pod, nil
			}
		}

	}
	return nil, nil
}
func (r *PodmigrationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	if err := mgr.GetFieldIndexer().IndexField(ctx, &corev1.Pod{}, podOwnerKey, func(raw runtime.Object) []string {
		pod := raw.(*corev1.Pod)
		owner := metav1.GetControllerOf(pod)
		if owner == nil {
			return nil
		}
		if owner.Kind != "Podmigration" {
			return nil
		}

		return []string{owner.Name}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&podmigv1.Podmigration{}).
		Complete(r)
}
