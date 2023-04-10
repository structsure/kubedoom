package entity

import (
	"fmt"
	"strings"
)

type Entity struct {
	Namespace string
	Pod       string
	Phase     string
}

func (entity Entity) ToPS() string {
	return fmt.Sprintf("%v/%v", entity.Namespace, entity.Pod)
}
func (entity Entity) ToNsAndPod() (string, string) {
	return entity.Namespace, entity.Pod
}
func (entity Entity) IsCurrently(inPhase string) Entity {
	if strings.Contains(entity.Phase, inPhase) {
		return entity
	}
	return Entity{}
}
func (entity Entity) Only(matching string) Entity {
	if strings.Contains(entity.Pod, matching) || strings.Contains(entity.Namespace, matching) {
		return entity
	}
	return Entity{}
}
func (entity Entity) Not(matching string) Entity {
	if !strings.Contains(entity.Pod, matching) && !strings.Contains(entity.Namespace, matching) {
		return entity
	}
	return Entity{}
}
