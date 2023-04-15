package mango

import (
	"strings"
)

func (cache *lru[K, V]) String() string {
	cache.lock.RLock()
	defer cache.lock.RUnlock()

	var sb strings.Builder
	node := cache.head.rhs
	sb.WriteString("MangoLRU [")
	for node != &cache.tail {
		sb.WriteString(node.String())
		sb.WriteRune(',')
		node = node.rhs
	}
	sb.WriteString("] (cache)")
	return sb.String()
}

func (cache *lru[K, V]) Evict(key K) (*V, bool) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	if n, ok := cache.nodes[key]; ok {
		cache.stats.EvictionExplicit.Count(1)
		delete(cache.nodes, key)
		cache.evict(n)
		return &n.value, true
	}
	return nil, false
}

func (cache *lru[K, V]) Clear() {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	cache.stats.Eviction.Count(uint64(len(cache.nodes)))

	cache.nodes = make(map[K]*lruNode[K, V], cache.maxSize)
	cache.head.rhs = &cache.tail
	cache.tail.lhs = &cache.head
}

func (cache *lru[K, V]) Stats() CacheStats {
	return cache.stats
}

func (cache *lru[K, V]) Get(key K) (*V, bool) {
	cache.lock.RLock()
	defer cache.lock.RUnlock()

	if n, ok := cache.nodes[key]; ok {
		cache.stats.Hit.Count(1)
		cache.recentlyUsed(n)
		return &n.value, true
	}
	cache.stats.Miss.Count(1)
	return nil, false
}

func (cache *lru[K, V]) Put(key K, value V) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	if n, ok := cache.nodes[key]; ok {
		cache.recentlyUsed(n)
		n.value = value
		return
	}

	n := newNode(key, value)
	n.lhs = &cache.head
	n.rhs = n.lhs.rhs
	n.rhs.lhs = n
	n.lhs.rhs = n
	cache.nodes[key] = n

	maxSize := cache.maxSize
	newSize := len(cache.nodes)
	if full := maxSize < newSize; !full {
		return
	}

	// can't be the head:
	eldest := cache.tail.lhs
	cache.stats.EvictionMaxSize.Count(1)
	delete(cache.nodes, eldest.key)
	cache.evict(eldest)
}

func (cache *lru[K, V]) recentlyUsed(node *lruNode[K, V]) {
	cache.evict(node)
	first := &cache.head
	third := first.rhs
	node.lhs = first
	node.rhs = third
	first.rhs = node
	third.lhs = node
}

func (cache *lru[K, V]) evict(node *lruNode[K, V]) {
	if next := node.rhs; next != nil {
		next.lhs = node.lhs
		node.lhs.rhs = next
	}
}
