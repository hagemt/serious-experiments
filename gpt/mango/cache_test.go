package mango

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	_ fmt.Stringer = &lru[string, int]{}
	_ fmt.Stringer = newNode("user0", 0)
)

type player struct {
	Score int `json:"player_score"`
}

func (p *player) String() string {
	v, _ := json.MarshalIndent(p, "", "\t")
	return string(v)
}

func TestStringers(t *testing.T) {
	s1 := newNode(1, 1).String()
	assert.Equal(t, "{1:1}", s1)
	s2 := NewLRU[string, string](CacheOptions{}).(*lru[string, string]).String()
	assert.Equal(t, "MangoLRU [] (cache)", s2)
	s3 := (&player{}).String()
	assert.Equal(t, strings.ReplaceAll("{'player_score':0}", "'", "\""), s3)
}

func TestLRU(t *testing.T) {
	lru := NewLRU[string, player](CacheOptions{
		MaxSize: 3,
	})
	require.Equal(t, 0, lru.Len())
	lru.Put("a", player{
		Score: 10,
	})
	require.Equal(t, 1, lru.Len())
	lru.Put("b", player{
		Score: 20,
	})
	require.Equal(t, 2, lru.Len())
	lru.Put("c", player{
		Score: 30,
	})
	require.Equal(t, 3, lru.Len())

	v, ok := lru.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 10, v.Score)

	lru.Put("d", player{
		Score: 40,
	})
	assert.Equal(t, 3, lru.Len())
	v, ok = lru.Get("b")
	assert.False(t, ok)
	assert.Nil(t, v)
}
