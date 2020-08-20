package g

import (
	"sync"
)

type Set struct {
	m map[string]bool
	sync.RWMutex
}

func SetNew() *Set {
	return &Set{
		m: map[string]bool{},
	}
}

func (this *Set) Add(item string) {
	this.Lock()
	defer this.Unlock()
	this.m[item] = true
}

func (this *Set) Remove(item string) {
	this.Lock()
	defer this.Unlock()
	delete(this.m, item)
}

func (this *Set) Has(item string) bool {
	this.RLock()
	defer this.RUnlock()
	_, ok := this.m[item]
	return ok
}

func (this *Set) Len() int {
	return len(this.List())
}

func (this *Set) Clear() {
	this.Lock()
	defer this.Unlock()
	this.m = map[string]bool{}
}

func (this *Set) List() []string {
	this.RLock()
	defer this.RUnlock()
	list := []string{}
	for item := range this.m {
		list = append(list, item)
	}
	return list
}