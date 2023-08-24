package kache

type PeerPicker interface {
	PickPeer(key string) (perr PeerGetter, ok bool)
}

type PeerGetter interface {
	Get(group, key string) ([]byte, error)
}
