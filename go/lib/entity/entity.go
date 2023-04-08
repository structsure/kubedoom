package entity

import "strings"

type Entity struct {
	ns     string
	pod    string
	status string
}

func (entity Entity) Only(matching string) Entity {
	if strings.Contains(entity.pod, matching) || strings.Contains(entity.ns, matching) {
		return entity
	}
	return Entity{}
}
func (entity Entity) Not(matching string) Entity {
	if !strings.Contains(entity.pod, matching) && !strings.Contains(entity.ns, matching) {
		return entity
	}
	return Entity{}
}
