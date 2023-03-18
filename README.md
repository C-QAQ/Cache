## Go语言从零实现分布式缓存

### cache/lru
- 设计并实现lru缓存淘汰规则

### cache/consistenthash
- 实现一致性哈希避免缓存雪崩

### cache/singleflight
- 解决并发请求相同缓存时的缓存击穿、缓存穿透问题

### cache/gocachepb
- 使用protobuf优化分布式节点间的信息通信效率