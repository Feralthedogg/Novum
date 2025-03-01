package state

type StateLayer struct {
	Counter int
}

func NewStateLayer(initial int) StateLayer {
	return StateLayer{Counter: initial}
}

func (s StateLayer) Increment() StateLayer {
	return StateLayer{Counter: s.Counter + 1}
}
