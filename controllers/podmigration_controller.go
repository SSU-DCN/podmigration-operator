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
	"os"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	podmigv1 "github.com/SSU-DCN/podmigration-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	podOwnerKey = ".metadata.controller"
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

	// Then list all pods controlled by the Podmigration resource object
	var childPods corev1.PodList
	if err := r.List(ctx, &childPods, client.InNamespace(req.Namespace), client.MatchingField(podOwnerKey, req.Name)); err != nil {
		log.Error(err, "unable to list child pods")
		return ctrl.Result{}, err
	}

	// First test log the number of pods
	size := len(childPods.Items)
	log.Info("", "template test", size)

	pod, err := r.desiredPod(migratingPod, &migratingPod, req.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}

	template := &migratingPod.Spec.Template
	annotations := getPodsAnnotationSet(template)
	// annotations := template.ObjectMeta
	log.Info("", "annotations ", annotations["snapshotPath"])
	log.Info("", "disired pod ", childPods)
	log.Info("", "disired pod ", pod)
	switch len(childPods.Items) {
	case 0:

		if annotations["snapshotPath"] == "" || annotations["snapshotPolicy"] == "" {
			log.Info("", "snapshotPolicy and snapshotPath is not given", annotations["snapshotPath"])
		} else {
			// snapshotPath and snapshotPolicy is given, should check if snapshotPath is exist or not
			_, err := os.Stat(annotations["snapshotPath"])
			if os.IsNotExist(err) {
				// if snapshotPath not found, delete snapshotPolicy and snapshotPath
				// Pod then start as normal
				pod.ObjectMeta.Annotations["snapshotPolicy"] = ""
				pod.ObjectMeta.Annotations["snapshotPath"] = ""
				log.Info("", "snapshotPath not found, we will start pod as normal", annotations["snapshotPath"])

			} else {
				// snapshotPath found, logging
				log.Info("", "snapshotPath found, we will start conatainer from checkpoint", annotations["snapshotPath"])
			}
		}
		if err := r.Create(ctx, pod); err != nil {
			log.Error(err, "unable to create Pod for MigratingPod", "pod", pod)
			return ctrl.Result{}, err
		}
	default:
		log.Info("", "no action", annotations["snapshotPath"])

	}
	return ctrl.Result{}, nil
}

func (r *PodmigrationReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(&corev1.Pod{}, podOwnerKey, func(raw runtime.Object) []string {
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
