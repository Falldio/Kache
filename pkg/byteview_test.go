package kache

import "testing"

func TestByteViewLen(t *testing.T) {
	bv := ByteView{bts: []byte("hello")}
	if bv.Len() != 5 {
		t.Fatal("bad length")
	}
}

func TestByteViewByteSlice(t *testing.T) {
	bv := ByteView{bts: []byte("hello")}
	if string(bv.ByteSlice()) != "hello" {
		t.Fatal("bad bytes")
	}
}

func TestByteViewString(t *testing.T) {
	bv := ByteView{bts: []byte("hello")}
	if bv.String() != "hello" {
		t.Fatal("bad string")
	}
}

func TestByteViewCloneBytes(t *testing.T) {
	bv := ByteView{bts: []byte("hello")}
	if string(cloneBytes(bv.bts)) != "hello" {
		t.Fatal("bad clone bytes")
	}
}
