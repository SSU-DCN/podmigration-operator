package endpoints

import (
	v1 "github.com/SSU-DCN/podmigration-operator/api/v1"
)

var From = &from{}

type from struct{}

func (c *from) Object(pm *v1.Podmigration) *Podmigration {
	return &Podmigration{
		Name:     pm.Name,
		Replicas: pm.Spec.Replicas,
		Selector: pm.Spec.Selector,
		// Template: pm.Spec.Template,
		Status: &pm.Status,
	}
}

func (c *from) List(list *v1.PodmigrationList) *List {
	items := make([]Podmigration, len(list.Items))
	for i, r := range list.Items {
		items[i] = *c.Object(&r)
	}
	return &List{
		Items: items,
	}
}
