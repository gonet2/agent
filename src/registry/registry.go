package registry

import (
	"sync"
)

import (
	. "types"
)

type Registry struct {
	records map[int32]chan IPCObject
	sync.RWMutex
}

var (
	_default_registry Registry
)

func init() {
	_default_registry.init()
}

func (r *Registry) init() {
	r.records = make(map[int32]chan IPCObject)
}

// register a user
func (r *Registry) Register(id int32, MQ chan IPCObject) {
	r.Lock()
	defer r.Unlock()
	r.records[id] = MQ
}

// unregister a user
func (r *Registry) Unregister(id int32) {
	r.Lock()
	defer r.Unlock()
	delete(r.records, id)
}

// deliver an object to the online user
func (r *Registry) Deliver(id int32, obj IPCObject) bool {
	r.RLock()
	defer r.RUnlock()
	if mq := r.records[id]; mq != nil {
		mq <- obj
		return true
	}
	return false
}

// return all online users
func (r *Registry) ListAll() (list []int32) {
	r.RLock()
	defer r.RUnlock()
	list = make([]int32, len(r.records))
	idx := 0
	for k := range r.records {
		list[idx] = k
		idx++
	}
	return
}

// return count of online users
func (r *Registry) Count() int {
	r.RLock()
	defer r.RUnlock()
	return len(r.records)
}

func Register(id int32, MQ chan IPCObject) {
	_default_registry.Register(id, MQ)
}

func Unregister(id int32) {
	_default_registry.Unregister(id)
}

func Deliver(id int32, obj IPCObject) bool {
	return _default_registry.Deliver(id, obj)
}

func ListAll() (list []int32) {
	return _default_registry.ListAll()
}

func Count() int {
	return _default_registry.Count()
}
