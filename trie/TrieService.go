/*
 * @Author: your name
 * @Date: 2020-04-02 15:06:27
 * @LastEditTime: 2020-04-07 16:11:50
 * @LastEditors: Please set LastEditors
 * @Description: In User Settings Edit
 * @FilePath: /Lemon/app/service/trie/TrieService.go
 */
package trie

import (
	"fmt"
	"strings"
	"sync"
)

// Trie s
type Trie struct {
	value    string         `json:"value"`
	children map[rune]*Trie `json:"children"`
	isEnd    bool           `json:"is_end"`
}

var root *Trie

// InitTrie s
func InitTrie() *Trie {
	root = new(Trie)
	root.children = make(map[rune]*Trie)
	root.isEnd = false

	return root
}

// Save 新增
func (t *Trie) Save(word string) {
	lock := &sync.RWMutex{}
	lock.Lock()
	defer lock.Unlock()

	for _, v := range word {
		if string(v) != "" && t.children[v] == nil {
			node := new(Trie)
			node.value = string(v)
			node.children = make(map[rune]*Trie)
			node.isEnd = false
			t.children[v] = node
			// t.isEnd = false
		}
		t = t.children[v]
	}
	t.isEnd = true
}

// Search 1
func (t *Trie) Search(word string) bool {
	word = strings.ReplaceAll(word, " ", "")
	for _, v := range word {
		if t.children[v] == nil {
			return false
		}
		t = t.children[v]
	}

	return t.isEnd
}

// SearchString 2
func (t *Trie) SearchString(prefix string) *Trie {
	// var target *Trie
	for _, v := range prefix {
		if t.children[v] == nil {
			// t.Save(string(v))
			return nil
		}
		// else {
		// 	target = t.children[v]
		// }
		t = t.children[v]
	}

	return t
}

// SearchPrefix 1
func (t *Trie) SearchPrefix(prefix string) *Trie {
	// var Strie *Trie
	Strie := t.SearchString(prefix)
	if Strie == nil {
		t.Save(prefix)
	}

	return Strie
}

// GetChildren 3
func (t *Trie) GetChildren(args ...interface{}) {
	var strArr *[]string
	if v, ok := args[0].(*[]string); ok {
		strArr = v
	}

	var text string = ""
	if len(args) > 1 {
		if v, ok := args[1].(string); ok {
			text = v
		}
	}

	if t != nil {
		if t.isEnd {
			*strArr = append(*strArr, text)
		}

		if len(t.children) > 0 {
			for _, node := range t.children {
				node.GetChildren(strArr, text+node.value)
			}
		}
	}
}

// GetOtherChildren 获得所有子节点并拼接
func (t *Trie) GetOtherChildren(args ...interface{}) {
	var strArr *[]string
	if v, ok := args[0].(*[]string); ok {
		strArr = v
	}

	// prefixTmp := ""
	// prefix := &prefixTmp
	// if len(args) > 1 {
	// 	if v, ok := args[1].(*string); ok {
	// 		prefix = v
	// 	}
	// }

	tmp := ""
	text := &tmp
	if len(args) > 1 {
		if v, ok := args[1].(*string); ok {
			text = v
		}
	}

	// *text = *prefix + t.value
	prefix := *text
	*text += t.value

	if t.isEnd {
		*strArr = append(*strArr, *text)
		if len(t.children) != 0 {
			// *prefix = *text
			*text = prefix
		}
	}

	for _, node := range t.children {
		fmt.Println(node)
		node.GetChildren(strArr, text)
	}
}
