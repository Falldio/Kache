package kache

import pb "github.com/falldio/Kache/pkg/proto/gen"

type PeerPicker interface {
	PickPeer(key string) (perr PeerGetter, ok bool)
}

type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}
