package xdb

import (
	"reflect"
	"sort"
	"strings"
)

func emptyFunc() {}

// Lock 加写锁
func Lock(r Model) bool {
	return lock(r, lockTypeWrite, false, false)
}

// Unlock 解写锁
func Unlock(r Model) {
	unlock(r, lockTypeWrite)
}

// RLock 加读锁
func RLock(r Model) bool {
	return lock(r, lockTypeRead, false, false)
}

// RUnlock 解读锁
func RUnlock(r Model) {
	unlock(r, lockTypeRead)
}

// LockIfAlive 如果存活则加写锁
func LockIfAlive(r Model) bool {
	return lock(r, lockTypeWrite, true, false)
}

// RLockIfAlive 如果存活则加读锁
func RLockIfAlive(r Model) bool {
	return lock(r, lockTypeRead, true, false)
}

// LockMulti 批量加写锁
func LockMulti(ms ...Model) func() {
	sort.Sort(LockSorter(ms))
	unlocked := make([]Model, 0, len(ms))
	var last Model
	for _, m := range ms {
		if last != m && lock(m, lockTypeWrite, false, false) {
			unlocked = append(unlocked, m)
			last = m
		}
	}

	return func() {
		for _, m := range unlocked {
			unlock(m, lockTypeWrite)
		}
	}
}

// RLockMulti 批量加读锁
func RLockMulti(ms ...Model) func() {
	sort.Sort(LockSorter(ms))
	unlocked := make([]Model, 0, len(ms))
	var last Model
	for _, m := range ms {
		if last != m && lock(m, lockTypeRead, false, false) {
			unlocked = append(unlocked, m)
			last = m
		}
	}

	return func() {
		for _, m := range unlocked {
			unlock(m, lockTypeRead)
		}
	}
}

func fLock(r Model) bool {
	return lock(r, lockTypeFree, false, false)
}

func fUnlock(r Model) {
	unlock(r, lockTypeFree)
}

func lock(m Model, lt lockType, onlyAlive bool, simulate bool) bool {
	// 在 actor 模型中不需要锁，直接返回 true
	if m == nil {
		return false
	}

	if onlyAlive && IsDeleted(m) {
		return false
	}

	return true
}

func unlock(m Model, lt lockType) {
	// 在 actor 模型中不需要锁，空操作
}

// LockSorter 锁排序器
type LockSorter []Model

func (ls LockSorter) Len() int {
	return len(ls)
}

func (ls LockSorter) Less(i, j int) bool {
	return compareLockPriority(ls[i], ls[j])
}

func (ls LockSorter) Swap(i, j int) {
	ls[i], ls[j] = ls[j], ls[i]
}

func compareLockPriority(last, curr Model) bool {
	// 在 actor 模型中不需要锁优先级，使用主键比较
	currType := reflect.TypeOf(curr)
	lastType := reflect.TypeOf(last)

	if currType == lastType {
		if last == curr {
			return true
		}
		// 直接使用主键比较
		return last.Source().PKComparator(PKOf(last), PKOf(curr)) < 0
	}

	// 不同类型使用命名空间比较
	ls := last.Source()
	cs := curr.Source()
	return strings.Compare(ls.Namespace, cs.Namespace) < 0
}

type lockType int8

const (
	lockTypeNone lockType = iota
	lockTypeRead
	lockTypeFree
	lockTypeWrite
)

func (lt lockType) String() string {
	switch lt {
	case lockTypeWrite:
		return "write"
	case lockTypeRead:
		return "read"
	case lockTypeFree:
		return "free"
	default:
		return "none"
	}
}

func (lt lockType) Lock(m Model, onlyAlive bool, simulate bool) bool {
	// 在 actor 模型中不需要锁，直接返回 true
	if m == nil {
		return false
	}

	if onlyAlive && IsDeleted(m) {
		return false
	}

	return true
}

func (lt lockType) Unlock(m Model) {
	// 在 actor 模型中不需要锁，空操作
}
