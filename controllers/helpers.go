package controllers

import (
	"fmt"

	podmigv1 "github.com/SSU-DCN/podmigration-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *PodmigrationReconciler) desiredPod(migratingPod podmigv1.Podmigration, parentObject runtime.Object, namespace string) (*corev1.Pod, error) {
	template := &migratingPod.Spec.Template
	desiredLabels := getPodsLabelSet(template)
	desiredFinalizers := getPodsFinalizers(template)
	desiredAnnotations := getPodsAnnotationSet(template)
	accessor, _ := meta.Accessor(parentObject)
	prefix := getPodsPrefix(accessor.GetName())

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace,
			Labels:       desiredLabels,
			Finalizers:   desiredFinalizers,
			Annotations:  desiredAnnotations,
			GenerateName: prefix,
		},
	}
	pod.Spec = *template.Spec.DeepCopy()
	if err := ctrl.SetControllerReference(&migratingPod, pod, r.Scheme); err != nil {
		return pod, err
	}
	return pod, nil
}

func getPodsLabelSet(template *corev1.PodTemplateSpec) labels.Set {
	desiredLabels := make(labels.Set)
	for k, v := range template.Labels {
		desiredLabels[k] = v
	}
	return desiredLabels
}

func getPodsFinalizers(template *corev1.PodTemplateSpec) []string {
	desiredFinalizers := make([]string, len(template.Finalizers))
	copy(desiredFinalizers, template.Finalizers)
	return desiredFinalizers
}

func getPodsAnnotationSet(template *corev1.PodTemplateSpec) labels.Set {
	desiredAnnotations := make(labels.Set)
	for k, v := range template.Annotations {
		desiredAnnotations[k] = v
	}
	return desiredAnnotations
}

func getPodsPrefix(controllerName string) string {
	// use the dash (if the name isn't too long) to make the pod name a bit prettier
	prefix := fmt.Sprintf("%s-", controllerName)
	// if len(validation.ValidatePodName(prefix, true)) != 0 {
	// 	prefix = controllerName
	// }
	return prefix
}

// RemoveString removes the element at position i from a string array without preserving
// the order.
// https://stackoverflow.com/a/37335777/4430124
func RemoveString(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
