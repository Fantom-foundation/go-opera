package piecefunc

const (
	// PercentUnit is used to define ratios with integers, it's 1.0
	PercentUnit = 1e6
)

// Dot is a pair of numbers
type Dot struct {
	X uint64
	Y uint64
}

// Mul is multiplication of ratios with integer numbers
func Mul(a, b uint64) uint64 {
	return a * b / PercentUnit
}

// Div is division of ratios with integer numbers
func Div(a, b uint64) uint64 {
	return a * PercentUnit / b
}

// Get calculates f(x), where f is a piecewise linear function defined by the pieces
func Get(x uint64, pieces []Dot) uint64 {

	// find a piece
	p0 := len(pieces) - 2
	for i, piece := range pieces {
		if i >= 1 && i < len(pieces)-1 && piece.X > x {
			p0 = i - 1
			break
		}
	}

	// linearly interpolate
	p1 := p0 + 1

	ratio := Div(x-pieces[p0].X, pieces[p1].X-pieces[p0].X)

	return Mul(pieces[p0].Y, (PercentUnit-ratio)) + Mul(pieces[p1].Y, ratio)
}
