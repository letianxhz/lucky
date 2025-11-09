package xdb

import (
	"sync"
)

// RepoOptions 仓库选项
type RepoOptions struct {
	GroupSize uint
}

// Repo 仓库（缓存）
type Repo struct {
	opts        RepoOptions
	initialized bool
	groups      []group
	groupMod    uint
	mu          sync.RWMutex
}

// Init 初始化仓库
func (r *Repo) Init(canonical bool, opts *RepoOptions, name string, keyComparator func(interface{}, interface{}) int, wg *sync.WaitGroup) {
	if opts != nil {
		r.opts = *opts
	}

	grpSize := r.opts.GroupSize
	if grpSize <= 0 {
		grpSize = 16
	}

	if r.initialized {
		panic("initialized cache")
	}

	r.opts.GroupSize = grpSize
	r.groupMod = grpSize - 1

	r.groups = make([]group, r.opts.GroupSize)
	for i := uint(0); i < r.opts.GroupSize; i++ {
		r.groups[i].Init(&r.opts, name, keyComparator, wg)
	}

	r.initialized = true
}

// Get 获取对象
func (r *Repo) Get(key Key) (interface{}, bool) {
	return r.getGroup(key).Get(key)
}

// SetOnFetch 在获取时设置
func (r *Repo) SetOnFetch(key Key, obj interface{}, volatile bool, callback func(obj any)) interface{} {
	return r.getGroup(key).SetOnFetch(key, obj, volatile, callback)
}

// SetOnStore 在存储时设置
func (r *Repo) SetOnStore(key Key, obj interface{}, volatile bool) bool {
	return r.getGroup(key).SetOnStore(key, obj, volatile)
}

// SetOnDelete 在删除时设置
func (r *Repo) SetOnDelete(key Key, obj interface{}) bool {
	return r.getGroup(key).SetOnDelete(key, obj)
}

// GetAll 获取所有匹配的对象
func (r *Repo) GetAll(key Key) ([]interface{}, bool) {
	return r.getGroup(key).GetAll(key)
}

// GetMulti 批量获取
func (r *Repo) GetMulti(keys []Key) ([]interface{}, []Key) {
	cached := make([]interface{}, 0, len(keys))
	remain := make([]Key, 0, len(keys))

	for _, key := range keys {
		if val, ok := r.Get(key); ok {
			cached = append(cached, val)
		} else {
			remain = append(remain, key)
		}
	}

	return cached, remain
}

// Expire 过期
func (r *Repo) Expire(key Key) {
	r.getGroup(key).Expire(key)
}

// Exists 检查是否存在
func (r *Repo) Exists(key Key) bool {
	_, ok := r.Get(key)
	return ok
}

// HashGroup 获取哈希组
func (r *Repo) HashGroup(key Key) int {
	return int(uint(key.HashGroup()) & r.groupMod)
}

func (r *Repo) getGroup(key Key) *group {
	return &r.groups[r.HashGroup(key)]
}

// group 组
type group struct {
	mu            sync.RWMutex
	data          map[Key]interface{}
	opts          *RepoOptions
	name          string
	keyComparator func(interface{}, interface{}) int
}

func (g *group) Init(opts *RepoOptions, name string, keyComparator func(interface{}, interface{}) int, wg *sync.WaitGroup) {
	g.data = make(map[Key]interface{})
	g.opts = opts
	g.name = name
	g.keyComparator = keyComparator
}

func (g *group) Get(key Key) (interface{}, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	val, ok := g.data[key]
	return val, ok
}

func (g *group) SetOnFetch(key Key, obj interface{}, volatile bool, callback func(obj any)) interface{} {
	g.mu.Lock()
	defer g.mu.Unlock()

	if existing, ok := g.data[key]; ok {
		return existing
	}

	if obj != nil {
		g.data[key] = obj
		if callback != nil {
			callback(obj)
		}
	} else {
		g.data[key] = nil
	}

	return obj
}

func (g *group) SetOnStore(key Key, obj interface{}, volatile bool) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.data[key]; exists {
		return false
	}

	g.data[key] = obj
	return true
}

func (g *group) SetOnDelete(key Key, obj interface{}) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.data[key]; !exists {
		return false
	}

	delete(g.data, key)
	return true
}

func (g *group) GetAll(key Key) ([]interface{}, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var results []interface{}
	for k, v := range g.data {
		if key.PrefixOf(k) {
			results = append(results, v)
		}
	}

	return results, len(results) > 0
}

func (g *group) Expire(key Key) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.data, key)
}
