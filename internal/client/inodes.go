package client

import (
	"github.com/s0rg/set"
)

type InodesMap struct {
	m map[string]map[int]set.Unordered[uint64]
	u map[string]map[int]set.Unordered[uint64]
	n map[string]map[int]string
	l map[string]map[int]string
}

type item struct {
	Cid string
	Pid int
}

func (m *InodesMap) AddProcess(containerID string, pid int, name string) {
	if m.n == nil {
		m.n = make(map[string]map[int]string)
	}

	names, ok := m.n[containerID]
	if !ok {
		names = make(map[int]string)
		m.n[containerID] = names
	}

	names[pid] = name
}

func (m *InodesMap) AddInode(containerID string, pid int, inode uint64) {
	if m.m == nil {
		m.m = make(map[string]map[int]set.Unordered[uint64])
	}

	pids, ok := m.m[containerID]
	if !ok {
		pids = make(map[int]set.Unordered[uint64])
		m.m[containerID] = pids
	}

	inodes, ok := pids[pid]
	if !ok {
		inodes = make(set.Unordered[uint64])
		pids[pid] = inodes
	}

	inodes.Add(inode)
}

func (m *InodesMap) MarkListener(containerID string, pid int, path string) {
	if m.l == nil {
		m.l = make(map[string]map[int]string)
	}

	pids, ok := m.l[containerID]
	if !ok {
		pids = make(map[int]string)
		m.l[containerID] = pids
	}

	pids[pid] = path
}

func (m *InodesMap) findListener(containerID string, pid int) (path string, ok bool) {
	if m.l == nil {
		return
	}

	pids, ok := m.l[containerID]
	if !ok {
		return
	}

	path, ok = pids[pid]

	return
}

func (m *InodesMap) nameFor(containerID string, pid int) (name string, ok bool) {
	if m.n == nil {
		return
	}

	names, ok := m.n[containerID]
	if !ok {
		return
	}

	name, ok = names[pid]

	return
}

func (m *InodesMap) MarkUnknown(containerID string, pid int, inode uint64) {
	if m.u == nil {
		m.u = make(map[string]map[int]set.Unordered[uint64])
	}

	pids, ok := m.u[containerID]
	if !ok {
		pids = make(map[int]set.Unordered[uint64])
		m.u[containerID] = pids
	}

	inodes, ok := pids[pid]
	if !ok {
		inodes = make(set.Unordered[uint64])
		pids[pid] = inodes
	}

	inodes.Add(inode)
}

func (m *InodesMap) ResolveUnknown(
	cb func(srcCID, dstCID, srcName, dstName, path string),
) {
	index := make(map[uint64]*item)

	for c, pids := range m.m {
		for p, inodes := range pids {
			inodes.Iter(func(k uint64) bool {
				index[k] = &item{
					Cid: c,
					Pid: p,
				}

				return true
			})
		}
	}

	for dstCID, dstPids := range m.u {
		for dstPID, inodes := range dstPids {
			inodes.Iter(func(k uint64) bool {
				known, ok := index[k]
				if !ok {
					return true
				}

				path, ok := m.findListener(known.Cid, known.Pid)
				if !ok {
					return true
				}

				srcName, ok := m.nameFor(known.Cid, known.Pid)
				if !ok {
					return true
				}

				dstName, ok := m.nameFor(dstCID, dstPID)
				if !ok {
					return true
				}

				cb(known.Cid, dstCID, srcName, dstName, path)

				return true
			})
		}
	}
}

func (m *InodesMap) Has(containerID string, pid int, inode uint64) (yes bool) {
	if m.m == nil {
		return false
	}

	pids, ok := m.m[containerID]
	if !ok {
		return false
	}

	inodes, ok := pids[pid]
	if !ok {
		return false
	}

	return inodes.Has(inode)
}
