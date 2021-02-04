package endpoints

import (
	"fmt"
	"strings"

	v1 "github.com/SSU-DCN/podmigration-operator/api/v1"
	"github.com/emicklei/go-restful"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	kubelog "sigs.k8s.io/controller-runtime/pkg/log"
)

// todo(TUONG): get namespace from request
var namespace = "default"

type PodmigrationEndpoint struct {
	client client.Client
}

func NewPodmigrationEndpoint(client client.Client) *PodmigrationEndpoint {
	return &PodmigrationEndpoint{client: client}
}

func (pe *PodmigrationEndpoint) SetupWithWS(ws *restful.WebService) {
	ws.Route(ws.GET("Podmigrations").To(pe.list).
		Doc("List of Podmigrations").
		Returns(200, "OK", &List{}))

	ws.Route(ws.POST("Podmigrations").To(pe.create).
		Doc("Create a new Podmigration").
		Reads(&Podmigration{}).
		Returns(200, "OK", &Podmigration{}).
		Returns(400, "Bad Request", nil))
}

func (pe *PodmigrationEndpoint) list(request *restful.Request, response *restful.Response) {
	dl := new(v1.PodmigrationList)
	err := pe.client.List(request.Request.Context(), dl, &client.ListOptions{})
	if err != nil {
		writeError(response, 404, Error{
			Title:   "Error",
			Details: fmt.Sprintf("Could not retrieve list: %s", err),
		})
	} else {
		l := From.List(dl)
		if err := response.WriteAsJson(l); err != nil {
			writeError(response, 404, Error{
				Title:   "Error",
				Details: "Could not list resources",
			})
		}
	}
}

func (pe *PodmigrationEndpoint) create(request *restful.Request, response *restful.Response) {
	pm := new(Podmigration)
	err := request.ReadEntity(pm)
	pm.Action = strings.ToLower(pm.Action)
	// fmt.Println("Calling an action: - %v", pm.Action)
	fmt.Println(pm)
	if err != nil {
		writeError(response, 400, Error{
			Title:   "Bad Request",
			Details: "Could not read entity",
		})
		return
	}

	if err := pm.Validate(); err != nil {
		writeError(response, 400, Error{
			Title:   "Validation error",
			Details: err.Error(),
		})
		return
	}

	// Check whether sourcePod of live-migration is exist or not
	var sourcePod *corev1.Pod
	var template corev1.PodTemplateSpec
	// if pm.Action == "live-migration" && pm.SourcePod != "" {
	if pm.SourcePod != "" {
		fmt.Println(pm.SourcePod)
		var childPods corev1.PodList
		if err := pe.client.List(request.Request.Context(), &childPods, client.InNamespace(namespace)); err != nil {
			writeError(response, 400, Error{
				Title:   "Bad Request",
				Details: "Could not find any running pod for migration",
			})
			return
		}

		if len(childPods.Items) > 0 {
			for _, pod := range childPods.Items {
				if pod.Name == pm.SourcePod && pod.Status.Phase == "Running" {
					sourcePod = pod.DeepCopy()
					container := sourcePod.Spec.Containers[0]
					template = corev1.PodTemplateSpec{
						ObjectMeta: sourcePod.ObjectMeta,
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  container.Name,
									Image: container.Image,
									Ports: container.Ports,
								},
							},
						},
					}
					if pm.DestHost != "" {
						template.Spec.NodeSelector = map[string]string{"kubernetes.io/hostname": pm.DestHost}
					}
				}
			}
		}
		if sourcePod == nil {
			writeError(response, 400, Error{
				Title:   "Bad Request",
				Details: "Could not find sourcePod for migration",
			})
			return
		}
	} else {
		template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"app": "redis"},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "redis",
						Image: "redis",
						Ports: []corev1.ContainerPort{
							{ContainerPort: 6379, Protocol: "TCP"},
						},
					},
				},
			},
		}
	}
	fmt.Println(template)
	// fmt.Println(pm.DestHost)
	// fmt.Println(template.Spec.NodeSelector)
	obj := &v1.Podmigration{
		ObjectMeta: metav1.ObjectMeta{Name: pm.Name, Namespace: "default"},
		Spec: v1.PodmigrationSpec{
			Replicas:     pm.Replicas,
			SourcePod:    pm.SourcePod,
			DestHost:     pm.DestHost,
			Selector:     pm.Selector,
			Action:       pm.Action,
			SnapshotPath: pm.SnapshotPath,
			Template:     template,
		},
	}
	err = pe.client.Create(request.Request.Context(), obj, &client.CreateOptions{})
	if err != nil {
		writeError(response, 400, Error{
			Title:   "Error",
			Details: fmt.Sprintf("Could not create object: %s", err),
		})
	} else {
		d := From.Object(obj)
		if err := response.WriteAsJson(d); err != nil {
			writeError(response, 422, Error{
				Title:   "Error",
				Details: "Could not write response",
			})
		}
	}
}

func writeError(response *restful.Response, httpStatus int, err Error) {
	if err := response.WriteHeaderAndJson(httpStatus, err, "application/json"); err != nil {
		kubelog.Log.Error(err, "Could not write the error response")
	}
}