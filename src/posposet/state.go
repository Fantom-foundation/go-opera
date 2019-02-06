package posposet

// State is a poset current state.
// TODO: make it internal.
type State struct {
	FrameIndex uint64
}

func (p *Poset) bootstrap() {
	// restore state
	p.state = p.store.GetState()
	if p.state == nil {
		p.state = &State{
			FrameIndex: 0,
		}
	}
	// TODO: restore all others from store.
}

// State saves current State
func (p *Poset) saveState() {
	p.store.SetState(p.state)
}
