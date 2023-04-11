package entity

import (
	"fmt"
	"log"
	"strings"
)

type Entity struct {
	Namespace string
	Pod       string
	Phase     string
}

func (e Entity) Log(message string) {
	log.Printf("%v: %v/%v while %v", message, e.Namespace, e.Pod, e.Phase)
}
func (entity Entity) ToPS() string {
	return fmt.Sprintf("%v/%v", entity.Namespace, entity.Pod)
}
func (entity Entity) ToNsAndPod() (string, string) {
	return entity.Namespace, entity.Pod
}
func (entity Entity) IsCurrently(inPhase string) Entity {
	entity.Log("Is Currently")
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
	var entities []Entity
	for _, matcher := range strings.Split(matchers, delmiter) {
		entities = append(entities, entity.Not(matcher))
	}
	return entities
}
