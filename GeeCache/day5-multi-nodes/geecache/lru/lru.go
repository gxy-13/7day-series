package lru

import "container/list"

// lru 最近最少使用， 维护一个双端队列， 如果某一个kv被访问了就移动到队尾，队首的就是将被淘汰的最近最少访问的记录

// 我们需要创建一个包含字典和双向链表的结构体类型Cache，方便实现CRUD

type Cache struct {
	maxBytes  int64                         // 最大内存
	nowBytes  int64                         // 当前已使用内存
	ll        *list.List                    // go标准库实现的双向列表
	cache     map[string]*list.Element      // map k 是string v是双向链表中的节点的指针
	onEvicted func(key string, value Value) // 某条记录被删除时的回掉函数，可以为nil
}

type entry struct {
	// 双向链表节点的类型
	key   string
	value Value
}

type Value interface {
	//Len 返回v所占空间的大小
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}

// Get 查找k所对应v在Cache中维护的双向链表中的节点，并且将节点移动队尾
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		// 如果在map中找到就队列中的entry移动到队尾，双向队列头尾是相对的，这里约定front就是尾
		c.ll.MoveToFront(ele)
		// 使用断言
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest 删除 就是将最近最少使用淘汰
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back() // front 约定为尾巴，back就是头
	if ele != nil {
		// 得到记录就从队列中删除
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		// 并将map中的记录也删除
		delete(c.cache, kv.key)
		// 将当前已使用的内存容量减去被淘汰的kv容量
		c.nowBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		// 如果有回掉函数就执行
		if c.onEvicted != nil {
			c.onEvicted(kv.key, kv.value)
		}
	}
}

// Add 如果k存在就是修改，不存在就是新增
func (c *Cache) Add(key string, value Value) {
	// 判断是否存在
	if ele, ok := c.cache[key]; ok {
		// 存在就是访问，移动到队尾
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		// 先修改已使用内存大小，再修改v的值
		c.nowBytes -= int64(kv.value.Len()) + int64(value.Len())
		kv.value = value
	} else {
		// 不存在就是新增
		// 将新的kv放到队尾
		ele := c.ll.PushFront(&entry{key, value})
		// 将ele加入map
		c.cache[key] = ele
		// 增加当前内存大小
		c.nowBytes += int64(len(key)) + int64(value.Len())
	}
	// 利用循环将超出最大内存的部分淘汰
	for c.nowBytes > c.maxBytes && c.maxBytes != 0 {
		c.RemoveOldest()
	}
}

// Len 获取双向队列中有多少数据
func (c *Cache) Len() int {
	return c.ll.Len()
}
