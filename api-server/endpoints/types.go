package endpoints

import (
	"errors"
	// "fmt"
	// "strings"

	v1 "github.com/SSU-DCN/podmigration-operator/api/v1"
	"github.com/emicklei/go-restful"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Endpoint interface {
	SetupWithWS(ws *restful.WebService)
}

type Podmigration struct {
	Name         string                 `json:"name"`
	DestHost     string                 `json:"destHost"`
	Replicas     int                    `json:"replicas"`
	Selector     *metav1.LabelSelector  `json:"selector"`
	Action       string                 `json:"action"`
	SnapshotPath string                 `json:"snapshotPath"`
	SourcePod    string                 `json:"sourcePod"`
	Template     corev1.PodTemplateSpec `json:"template,omitempty"`
	Status       *v1.PodmigrationStatus `json:"status,omitempty"`
}

func (pm *Podmigration) Validate() error {
	var validated bool
	validated = true
	//TODO(Tuong): check template is valid or not
	// if pm.Template == checkTemplate {
	// 	return error.New("template can't be empty")
	// } else {
	// 	validated = true
	// }
	if validated {
		return nil
	}
	return errors.New("source type validation was not performed, type can only be [WebFolder,S3]")
}

type List struct {
	Items []Podmigration `json:"items"`
}

type Error struct {
	Title   string `json:"title"`
	Details string `json:"details"`
}
