package objectRef

import (
	"fmt"
	"maps"
)

const (
	_          int = iota
	LITERAL        // 1
	FIELD          // 2
	EXPRESSION     // 3
	GROUP          // 4
)

type LocalReferences struct {
	store map[string]int
}

func (lr *LocalReferences) Set(id string, ref int) {
	lr.store[id] = ref
}

func (lr *LocalReferences) Get(id string) int {
	return lr.store[id]
}

func NewLocalReferences() *LocalReferences {
	s := make(map[string]int)
	return &LocalReferences{store: s}
}

func (lr *LocalReferences) Highest() int {
	output := -1
	for _, i := range lr.store {
		if i > output {
			output = i
		}
	}

	return output
}

func (lr *LocalReferences) AllSame() bool {
	h := lr.Highest()

	if h == -1 {
		return true
	}

	for _, i := range lr.store {
		if i != h {
			return false
		}
	}

	return true
}

func (lr *LocalReferences) List() []string {
	list := []string{}

	for i, v := range lr.store {
		list = append(list, fmt.Sprintf("%s %d", i, v))
	}

	return list
}

func (lr *LocalReferences) Append(subRef *LocalReferences) {
	maps.Copy(lr.store, subRef.store)
}
