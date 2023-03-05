package lru

import "container/list"

type Cache struct {
	maxBytes         int64
	nowBytes         int64
	DoubleLinkedList *list.List
	Cache            map[string]*list.Element
	OnEvicted        func(key string, value Value)
}

type Value interface {
	Len() int64
}

type Entry struct {
	Key   string
	Value Value
}

func NewCache(capacity int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:         capacity,
		nowBytes:         0,
		DoubleLinkedList: list.New(),
		Cache:            make(map[string]*list.Element, capacity),
		OnEvicted:        onEvicted,
	}
}

func (cache *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := cache.Cache[key]; ok {
		cache.DoubleLinkedList.MoveToBack(ele)
		kv := ele.Value.(*Entry)
		return kv.Value, true
	}
	return
}

func (cache *Cache) RemoveOldest() {
	ele := cache.DoubleLinkedList.Front()
	if ele != nil {
		cache.DoubleLinkedList.Remove(ele)
		kv := ele.Value.(*Entry)
		delete(cache.Cache, kv.Key)
		cache.nowBytes = cache.nowBytes - int64(len(kv.Key)) - kv.Value.Len()
		if cache.OnEvicted != nil {
			cache.OnEvicted(kv.Key, kv.Value)
		}
	}
}

func (cache *Cache) Add(key string, value Value) {
	if ele, ok := cache.Cache[key]; ok {
		kv := ele.Value.(*Entry)
		kv.Value = value
		cache.DoubleLinkedList.MoveToBack(ele)
		cache.nowBytes += cache.nowBytes + int64(len(kv.Key)) + kv.Value.Len()
	} else {
		ele := cache.DoubleLinkedList.PushBack(&Entry{Key: key, Value: value})
		cache.Cache[key] = ele
		cache.nowBytes = cache.nowBytes + int64(len(key)) + value.Len()
	}
	for cache.maxBytes != 0 && cache.nowBytes < cache.maxBytes {
		cache.RemoveOldest()
	}
}
