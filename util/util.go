package util

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

//声明一个结构体 数据项
type Item struct {
	Object     interface{} //真正保存的数据,存储任意类型的数据对象
	Expiration int64       //数据保留的时间,该数据项的过期时间
}

//定义过期的常量
const (
	//没有过期的标志
	NoExpiration time.Duration = -1

	//默认的过期时间
	DefaultExpiration time.Duration = 0
)

//声明一个缓存的结构体，保存基本的信息
type Cache struct {
	defaultExpiration time.Duration
	items             map[string]Item //缓存数据项存储在map中
	mu                sync.RWMutex    //读写锁
	gcInterval        time.Duration   //过期数据项清理的周期
	stopGc            chan bool
}

//Exipred 判断是否过期 返回true 或者false
func (item Item) Expired() bool {
	//如果没有设置的话，就返回为空
	if item.Expiration == 0 {
		return false
	}
	//否则就是直接用当前的时间与过期的时间作比较
	return time.Now().UnixNano() > item.Expiration
}

//过期缓存数据项清理
func (c *Cache) gcLoop() {
	//通过GC来回收垃圾数据
	ticker := time.NewTicker(c.gcInterval)
	//不断的去循环通道里的数据，通过 time.Ticker 定期执行 DeleteExpired() 方法
	for {
		select {
		//指定的c.Interval 间隔时间
		case <-ticker.C:
			//清理过期的数据项
			c.DeleteExpired()
		case <-c.stopGc:
			//通过监听c.stopGc管道，如果有数据从该管道中发送过来，我们就停止 gcLoop() 的运行
			ticker.Stop()
			//直接return掉
			return
		}
	}
}

//删除缓存的数据项
func (c *Cache) delete(k string) {
	delete(c.items, k)
}

//删除过期的数据项
func (c *Cache) DeleteExpired() {
	//获取当前的时间戳
	now := time.Now().UnixNano()
	//加上锁的机制
	c.mu.Lock()
	//延迟解锁
	defer c.mu.Unlock()
	//遍历所有的数据项
	for k, v := range c.items {
		//判断当前的时间戳，如果当前的时间戳大于过期的时间戳，就删除这个数据项
		if v.Expiration > 0 && now > v.Expiration {
			//删除这个数据项
			c.delete(k)
		}
	}
}

//设置缓存数据项，set,如果数据存在，那么直接覆盖
//需要传入数据的name值，以及value值，以及过期的时间
func (c *Cache) Set(k string, v interface{}, d time.Duration) {
	var e int64
	//设置永不过期
	if d == NoExpiration {
		d = 0
	}
	//默认的过期时间
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	//如果过期的时间大于0，那么直接添加
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}
	//设置特定的数据项，切片
	c.items[k] = Item{
		Object:     v,
		Expiration: e,
	}
	var filepath = "C:\\tmp\\MyRedis\\set.txt"
	c.SaveToFile(filepath)
}

// 将缓存数据项写入到 io.Writer 中
func (c *Cache) Save(w io.Writer) (err error) {
	enc := gob.NewEncoder(w)
	defer func() {
		if x := recover(); x != nil {
			err = fmt.Errorf("Error registering item types with Gob library")
		}
	}()
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, v := range c.items {
		gob.Register(v.Object)
	}
	err = enc.Encode(&c.items)
	return
}

// 从 io.Reader 中读取数据项
func (c *Cache) Load(r io.Reader) error {
	dec := gob.NewDecoder(r)
	items := map[string]Item{}
	err := dec.Decode(&items)
	if err == nil {
		c.mu.Lock()
		defer c.mu.Unlock()
		for k, v := range items {
			ov, found := c.items[k]
			if !found || ov.Expired() {
				c.items[k] = v
			} else {
				//只有没有过期的数据才打印
				fmt.Println("The key ", k, " of value is:", c.items[k].Object)
			}
		}
	}
	return err
}

// 从文件中加载缓存数据项
func (c *Cache) LoadFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	if err = c.Load(f); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// 保存数据项到文件中
func (c *Cache) SaveToFile(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	if err = c.Save(f); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

//获得数据项，如果找到了数据项，还需要判断数据项是否已经过期了
//返回值是数据的值，以及true/false
func (c *Cache) get(k string) (interface{}, bool) {
	item, found := c.items[k]
	if !found {
		//没有找到数据的话，直接返回空
		return nil, false
	}
	if item.Expired() {
		//如果发现数据过期了，直接返回空
		return nil, false
	}
	//否则直接返回数据
	return item.Object, true
}

func (c *Cache) All() {
	//否则直接返回数据
	for key := range c.items {
		fmt.Printf("The key %s is:%v", key, c.items[key].Object)
		fmt.Println()
	}
	// return c.items, true
}

//如果发现数据项已经存在则会发生报错，这样能避免缓存被错误的覆盖
//添加数据项，如果数据项已经存在，则返回错误
func (c *Cache) Add(k string, v interface{}, d time.Duration) error {
	//添加之前先加锁
	c.mu.Lock()
	_, found := c.get(k)
	if found {
		c.mu.Unlock()
		//如果发现数据项早就存在，直接返回错误信息
		return fmt.Errorf("Item %s already exists", k)

	}
	//添加成功之后，错误就为空
	c.Set(k, v, d)
	c.mu.Unlock()
	return nil

}

//传入name值，反正数据的value，以及bool值
func (c *Cache) Get(k string) (interface{}, bool) {
	//先加锁
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[k]
	if !found {
		return nil, false

	}
	if item.Expired() {
		return nil, false
	}
	//直接返回数据项
	return item.Object, true
}

//替换一个存在的数据项
//传入name值，value值，以及过期的时间
//返回一个error
func (c *Cache) Replace(k string, v interface{}, d time.Duration) error {
	c.mu.Lock()
	_, found := c.get(k)
	if !found {
		c.mu.Unlock()
		return fmt.Errorf("Item %s doesn't exist", k)
	}
	c.Set(k, v, d)
	c.mu.Unlock()
	return nil
}

//删除一个数据项
func (c *Cache) Delete(k string) {
	//加锁
	c.mu.Lock()
	c.delete(k)
	//释放锁
	c.mu.Unlock()
}

// 创建一个缓存系统
func NewCache(defaultExpiration, gcInterval time.Duration) *Cache {
	c := &Cache{
		defaultExpiration: defaultExpiration,
		gcInterval:        gcInterval,
		items:             map[string]Item{},
		stopGc:            make(chan bool),
	}
	// 开始启动过期清理 goroutine
	go c.gcLoop()
	return c
}
