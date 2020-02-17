package csimanager

import (
	"sync"

	"github.com/hashicorp/nomad/nomad/structs"
)

// volumeUsageTracker tracks the allocations that depend on a given volume
type volumeUsageTracker struct {
	state   map[volumeUsageKey][]*structs.Allocation
	stateMu sync.Mutex
}

func newVolumeUsageTracker() *volumeUsageTracker {
	return &volumeUsageTracker{
		state: make(map[volumeUsageKey][]*structs.Allocation),
	}
}

type volumeUsageKey struct {
	volume *structs.CSIVolume
	// usageOpts *UsageOptions
}

func (v *volumeUsageTracker) allocsForKey(key volumeUsageKey) []*structs.Allocation {
	return v.state[key]
}

func (v *volumeUsageTracker) appendAlloc(key volumeUsageKey, alloc *structs.Allocation) {
	allocs := v.allocsForKey(key)
	allocs = append(allocs, alloc.CopySkipJob())
	v.state[key] = allocs
}

func (v *volumeUsageTracker) removeAlloc(key volumeUsageKey, needle *structs.Allocation) {
	allocs := v.allocsForKey(key)
	var newAllocs []*structs.Allocation
	for _, alloc := range allocs {
		if alloc.ID != needle.ID {
			newAllocs = append(newAllocs, alloc)
		}
	}

	if len(newAllocs) == 0 {
		delete(v.state, key)
	} else {
		v.state[key] = newAllocs
	}
}

func (v *volumeUsageTracker) Claim(alloc *structs.Allocation, volume *structs.CSIVolume /*, usage *UsageOptions*/) {
	v.stateMu.Lock()
	defer v.stateMu.Unlock()

	key := volumeUsageKey{volume: volume /*, usageOpts: usage*/}
	v.appendAlloc(key, alloc)
}

// Free removes the allocation from the state list for the given alloc. If the
// alloc is the last allocation for the volume then
func (v *volumeUsageTracker) Free(alloc *structs.Allocation, volume *structs.CSIVolume /*, usage *UsageOptions*/) bool {
	v.stateMu.Lock()
	defer v.stateMu.Unlock()

	key := volumeUsageKey{volume: volume /*, usageOpts: usage*/}
	v.removeAlloc(key, alloc)
	allocs := v.allocsForKey(key)
	return len(allocs) == 0
}
