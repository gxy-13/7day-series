package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// consistentHash 一致性哈希   走向分布式哈希的重要环节，
// 假设有十个节点，第一次访问选择节点1，节点1从数据源获取并缓存，第二次访问，有9/10的几率访问的节点没有上次访问的数据，需要再去数据源获取，效率低

/*
   1.将key和节点映射到0到2^32-1首尾相连的环上，key顺时针寻找到第一个节点就是应取的节点/机器。这样在新增/删除节点时，
	只需要重新定位该节点附近的一小部分数据。
   2.设置虚拟节点解决了当节点过少时出现的数据倾斜问题，用map来维护虚拟节点和真实节点
*/

type Hash func(data []byte) uint32

type Map struct {
	hash     Hash           // Hash函数
	replicas int            // 虚拟节点倍数
	hashMap  map[int]string // 虚拟节点和真实节点映射
	keys     []int          // 哈希环
}

func NewMap(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	// 没有指定哈希函数，设置默认
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 添加真实节点，允许传入多个或0个真实节点的名称
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		// 对于每个真实节点，创建m.replicas个虚拟节点
		for i := 0; i < m.replicas; i++ {
			// 每一个虚拟节点的名称是数字编号加key
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			// 将虚拟节点映射到环上
			m.keys = append(m.keys, hash)
			// 添加虚拟和真实节点映射关系
			m.hashMap[hash] = key
		}
	}
	// 将环上的哈希值排序
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	// 计算key的hash
	hash := int(m.hash([]byte(key)))
	// 顺时针找到第一个匹配的虚拟节点的下表
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	// 因为是一个哈希环，所以用取余来获取
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
