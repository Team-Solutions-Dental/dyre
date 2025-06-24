package objectRef

const (
	_ int = iota
	LITERAL
	FIELD
	EXPRESSION
	GROUP
)

type LocalReferences struct {
	store map[string]int
}

func (lr *LocalReferences) Set(id string, ref int) {
	lr.store[id] = ref
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
