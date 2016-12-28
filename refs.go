package systray

import (
	"math/rand"
	"sync"
	"sync/atomic"
)

/*
#include <inttypes.h>
*/
import "C"

/*
Storage implementation borrowed from:
https://github.com/justinfx/gofileseq/blob/master/cpp/export/storage.go
*/

func init() {
	gSystrays = &systrayMap{
		lock: new(sync.RWMutex),
		m:    make(map[refId]*systrayRef),
	}
}

// global map to track references to _Systray instances
var gSystrays *systrayMap

type refId uint64

// uuid generates a new unique id using the shared generator
func uuid() refId {
	u := uint64(rand.Uint32())<<32 + uint64(rand.Uint32())
	return refId(u)
}

// systrayRef is a reference entry for a tracked _Systray instance
// in the map
type systrayRef struct {
	*_Systray
	refs uint32
}

// systrayMap tracks references to _Systray instances,
// and provides reference counting and lookup via an id.
type systrayMap struct {
	lock *sync.RWMutex
	m    map[refId]*systrayRef
}

// Len returns the number of _Systray references that are being tracked
func (m *systrayMap) Len() int {
	m.lock.RLock()
	l := len(m.m)
	m.lock.RUnlock()
	return l
}

// Add adds a new references to a _Systray instance, and increments
// the reference count to 1
func (m *systrayMap) Add(tray *_Systray) refId {
	m.lock.Lock()
	id := uuid()
	m.m[id] = &systrayRef{tray, 1}
	m.lock.Unlock()
	// fmt.Printf("storage: Added ref to %v with id %d\n", tray, id)
	return id
}

// Incref increments the reference count for a tracked reference
func (m *systrayMap) Incref(id refId) {
	m.lock.RLock()
	ref, ok := m.m[id]
	m.lock.RUnlock()

	if !ok {
		return
	}

	atomic.AddUint32(&ref.refs, 1)
	// fmt.Printf("storage: Incref %v (id: %v) to %d\n", ref, id, ref.refs)
}

// Decref decrements a reference count for a tracked reference.
// If the count reaches zero, it will be removed from the map.
func (m *systrayMap) Decref(id refId) {
	m.lock.RLock()
	ref, ok := m.m[id]
	m.lock.RUnlock()

	if !ok {
		return
	}

	refs := atomic.AddUint32(&ref.refs, ^uint32(0))
	// fmt.Printf("storage: Decref %v to %d\n", ref, refs)
	if refs != 0 {
		return
	}

	m.lock.Lock()
	if atomic.LoadUint32(&ref.refs) == 0 {
		// fmt.Printf("storage: Deleting %v\n", ref)
		delete(m.m, id)
	}
	m.lock.Unlock()
}

// Get returns an tracked _Systray reference by its id,
// and a bool indicating whether the reference exists.
// If the reference does not exist, then the pointer will
// be invalid
func (m *systrayMap) Get(id refId) (*systrayRef, bool) {
	m.lock.RLock()
	ref, ok := m.m[id]
	m.lock.RUnlock()
	return ref, ok
}

// Exported functions for C
//

//export systray_Incref
func systray_Incref(id refId) {
	gSystrays.Incref(id)
}

//export systray_Decref
func systray_Decref(id refId) {
	gSystrays.Decref(id)
}
