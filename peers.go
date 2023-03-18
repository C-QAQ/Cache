package gocache

import (
	pb "gocache/gocachepb"
)

// PeerPicker 根据传入的key选择相应节点PeerGetter
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter 从对应group查找缓存值
type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}
