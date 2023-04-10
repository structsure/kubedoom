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
func (entity Entity) Only(matcher string) Entity {
	if strings.Contains(entity.Pod, matcher) || strings.Contains(entity.Namespace, matcher) {
		return entity
	}
	return Entity{}
}
func (entity Entity) Not(matcher string) Entity {
	if !strings.Contains(entity.Pod, matcher) && !strings.Contains(entity.Namespace, matcher) {
		return entity
	}
	return Entity{}
}
func (entity Entity) NotList(matchers, delmiter string) []Entity {
	entities := []Entites{}
	for _, matcher := range strings.Split(matchers, delmiter) {
		append(attenties, entity.Not(matcher))
	}
	return entities
}
