package gocachex

import pb "goCacheX/gocacheXpb"

// PeerPicker is the interface that must be implemented to locate
// the peer that owns a specific key.
// 选择节点
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter is the interface that must be implemented by a peer.
// 获取节点
type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}
