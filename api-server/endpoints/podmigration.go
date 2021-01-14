package endpoints

import (
	"fmt"

	v1 "github.com/SSU-DCN/podmigration-operator/api/v1"
	"github.com/emicklei/go-restful"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	kubelog "sigs.k8s.io/controller-runtime/pkg/log"
)

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

	//TODO(tuong): convert request data-raw to this Template.
	template := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      map[string]string{"app": "redis"},
			Annotations: map[string]string{"snapshotPolicy": "restore", "snapshotPath": "/var/lib/kubelet/migration/" + pm.Name},
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

	obj := &v1.Podmigration{
		ObjectMeta: metav1.ObjectMeta{Name: pm.Name, Namespace: "default"},
		Spec:       v1.PodmigrationSpec{Replicas: pm.Replicas, Selector: pm.Selector, Template: template},
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
