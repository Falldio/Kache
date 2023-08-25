package kache

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

var mockGetter Getter = GetterFunc(func(key string) ([]byte, error) {
	return []byte(key), nil
})

type MockPeer struct {
	mock.Mock
}

func (m *MockPeer) PickPeer(key string) (PeerGetter, bool) {
	args := m.Called(key)
	return args.Get(0).(PeerGetter), args.Bool(1)
}

func (m *MockPeer) Update(group, key string, value []byte) error {
	args := m.Called(group, key, value)
	return args.Error(0)
}

type MockPeerGetter struct {
	mock.Mock
}

func (m *MockPeerGetter) Get(group, key string) ([]byte, error) {
	args := m.Called(group, key)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockPeerGetter) Watch(group, key string, fn func([]byte)) {
	fn([]byte(key))
	m.Called(group, key, fn)
}

func TestNewGroup(t *testing.T) {
	assert.Panics(t, func() { NewGroup("", 2<<10, nil) })

	g := NewGroup("scores", 2<<10, mockGetter)
	assert.Equal(t, "scores", g.name)
	assert.Equal(t, int64(2<<10), g.cacheBytes)
	assert.Equal(t, groups["scores"], g)
}

func TestGetter(t *testing.T) {
	expect := []byte("key")
	if v, _ := mockGetter.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Fatalf("callback failed, expect: %v, got: %v", expect, v)
	}
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	g := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key]++
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s does not exits", key)
		}))

	for k := range db {
		if _, err := g.Get(k); err != nil {
			t.Fatalf("failed to get value of %v", k)
		}
		if _, err := g.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s missing", k)
		}
	}

	if view, err := g.Get("unknown"); err == nil {
		t.Fatalf("expect empty, get %s", view)
	}

	if _, err := g.Get(""); err == nil {
		t.Fatalf("expect empty, get %s", "")
	}
}

func TestGetGroup(t *testing.T) {
	g := NewGroup("scores", 2<<10, mockGetter)
	g1 := GetGroup("scores")
	assert.Equal(t, g, g1)
}

func TestSet(t *testing.T) {
	g := NewGroup("scores", 2<<10, mockGetter)
	g.Set("Tom", []byte("630"), 0)
	if v, err := g.Get("Tom"); err != nil || string(v.bts) != "630" {
		t.Fatalf("failed to get value of %v", "Tom")
	}
	assert.Equal(t, false, g.Set("", []byte("630"), 0))
}

func TestLookupCache(t *testing.T) {
	// cacheBytes <= 0
	g := NewGroup("scores", 0, mockGetter)
	g.Set("Tom", []byte("630"), 0)
	v, ok := g.lookupCache("Tom")
	assert.Equal(t, false, ok)
	assert.Equal(t, "", string(v.bts))

	// hotCache hit
	g = NewGroup("scores", 2<<10, mockGetter)
	g.hotCache.Set("Tom", ByteView{bts: []byte("630")}, 0)
	v, ok = g.lookupCache("Tom")
	assert.Equal(t, true, ok)
	assert.Equal(t, "630", string(v.bts))
}

func TestRegisterPeers(t *testing.T) {
	g := NewGroup("scores", 2<<10, mockGetter)
	m := &MockPeer{}
	g.RegisterPeers(m)
	assert.Equal(t, m, g.peers)
	assert.Panics(t, func() { g.RegisterPeers(m) })
}

func TestLoad(t *testing.T) {
	mockPeer := &MockPeer{}
	mockPeerGetter := &MockPeerGetter{}
	mockPeer.On("PickPeer", "Tom").Return(mockPeerGetter, true)
	mockPeerGetter.On("Watch", "scores", "Tom", mock.AnythingOfType("func([]uint8)")).Return()
	mockCall := mockPeerGetter.On("Get", "scores", "Tom").Return([]byte("630"), nil)
	g := NewGroup("scores", 2<<10, mockGetter)
	g.RegisterPeers(mockPeer)
	g.load("Tom")
	mockPeer.AssertCalled(t, "PickPeer", "Tom")
	mockPeerGetter.AssertCalled(t, "Get", "scores", "Tom")

	mockCall.Unset()
	mockPeerGetter.On("Get", "scores", "Tom").Return([]byte{}, fmt.Errorf("not found"))
	g.load("Tom")
	mockPeerGetter.AssertCalled(t, "Get", "scores", "Tom")
}

func TestGetFromPeer(t *testing.T) {
	mockPeerGetter := &MockPeerGetter{}
	mockPeerGetter.On("Get", "scores", "Tom").Return([]byte("630"), nil)
	mockPeerGetter.On("Watch", "scores", "Tom", mock.AnythingOfType("func([]uint8)")).Return()
	g := NewGroup("scores", 2<<10, mockGetter)
	g.getFromPeer(mockPeerGetter, "Tom")
	mockPeerGetter.AssertCalled(t, "Get", "scores", "Tom")
}

func TestPopulateCache(t *testing.T) {
	g := NewGroup("scores", 0, mockGetter)
	g.populateCache("Tom", ByteView{bts: []byte("630")}, &g.hotCache)

	// trigger cache replacement
	g = NewGroup("scores", 32, mockGetter)

	for i := 0; i < 8; i++ {
		g.mainCache.Set(fmt.Sprintf("%d", i), ByteView{bts: []byte("0")}, 0)
		g.hotCache.Set(fmt.Sprintf("%d", i), ByteView{bts: []byte("0")}, 0)
	}
	assert.Equal(t, 8, g.mainCache.Len())
	assert.Equal(t, 8, g.hotCache.Len())
	g.populateCache("Tom", ByteView{bts: []byte("630")}, &g.hotCache)
	assert.LessOrEqual(t, int64(2), g.hotCache.Bytes())
}
