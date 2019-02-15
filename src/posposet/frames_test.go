package posposet

/*
 * Poset's test methods:
 */

// RootFrame returns frameN of root event.
// Returns nil if event is not root or frame is too old.
// It is for test purpose only.
func (p *Poset) RootFrame(root *Event) *uint64 {
	for n, f := range p.frames {
		if roots := f.NodeRootsGet(root.Creator); roots != nil {
			if hashes := roots[root.Creator]; hashes != nil {
				if hashes.Contains(root.Hash()) {
					return &n
				}
			}
		}
	}
	return nil
}
