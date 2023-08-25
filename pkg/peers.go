package kache

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
	Update(group, key string, value []byte) error
}

type PeerGetter interface {
	Get(group, key string) ([]byte, error)
	Watch(group, key string, fn func([]byte))
}
