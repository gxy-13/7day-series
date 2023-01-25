package geecache

import (
	"fmt"
	"log"
	"sync"
)

// 主结构体Group， 负责用户的交互，并且控制缓存存储和获取的流程

/*
	接收key ----- 检查是否被缓存 ----Y---- 获取缓存值 （1）
					｜
					｜---N--- 是否应当从远程节点获取 ---Y---- 与远程节点交互 ---- 返回缓存值 (2)
									｜
									｜-----N---- 调用回调函数，获取值并添加到缓存---- 返回缓存值 （3）
*/

// Getter 当缓存不存在，应该要从数据源（文件，数据库）中获取并添加到数据库，我们不应该支持多种数据源配置，种类太多，扩展差。如何从源头获取，让用户决定
// 设计一个回调函数，当缓存不存在是，调用这个函数，得到源数据
type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

// Get 定义一个函数类型F，实现接口A的方法，在方法中调用自己，这是GO中将其他函数(参数返回值与F一致)转换为接口A的常用技巧
func (f GetterFunc) Get(key string) ([]byte, error) {
	// f 是一个函数， 将key作为其参数传入
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

type Group struct {
	name      string //name是唯一命名空间 比如创建三个group，学生成绩为scores，信息为info
	getter    Getter // 回调函数，缓存未命中获取源数据的回调
	mainCache cache  // 刚刚实现的并发缓存
}

func NewGroup(name string, getter Getter, cacheBytes int64) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:   name,
		getter: getter,
		mainCache: cache{
			cacheBytes: cacheBytes,
		},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	// 流程1 从缓存中寻找是否存在，存在返回，不存在调用load方法
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	// 调用回调函数 获取源数据，并添加到mainCache中
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneByte(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
