package kache

// A ByteView holds an immutable view of bytes
type ByteView struct {
	bts []byte
}

func (bv ByteView) Len() int {
	return len(bv.bts)
}

func (bv ByteView) ByteSlice() []byte {
	return cloneBytes(bv.bts)
}

func (bv ByteView) String() string {
	return string(bv.bts)
}

func cloneBytes(bts []byte) []byte {
	c := make([]byte, len(bts))
	copy(c, bts)
	return c
}
