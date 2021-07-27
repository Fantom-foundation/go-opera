package piecefunc

const (
	// DecimalUnit is used to define ratios with integers, it's 1.0
	DecimalUnit = 1e6
)

// Dot is a pair of numbers
type Dot struct {
	X uint64
	Y uint64
}

type Func struct {
	dots []Dot
}

func NewFunc(dots []Dot) func(x uint64) uint64 {
	if len(dots) < 2 {
		panic("too few dots")
	}

	var prevX uint64
	for i, dot := range dots {
		if i >= 1 && dot.X <= prevX {
			panic("non monotonic X")
		}
		prevX = dot.X
	}

	return Func{
		dots: dots,
	}.Get
}

// Mul is multiplication of ratios with integer numbers
func Mul(a, b uint64) uint64 {
	return a * b / DecimalUnit
}

// Div is division of ratios with integer numbers
func Div(a, b uint64) uint64 {
	return a * DecimalUnit / b
}

// Get calculates f(x), where f is a piecewise linear function defined by the pieces
func (f Func) Get(x uint64) uint64 {
	// find a piece
	p0 := len(f.dots) - 2
	for i, piece := range f.dots {
		if i >= 1 && i < len(f.dots)-1 && piece.X > x {
			p0 = i - 1
			break
		}
	}
	// linearly interpolate
	p1 := p0 + 1

	x0, x1 := f.dots[p0].X, f.dots[p1].X
	y0, y1 := f.dots[p0].Y, f.dots[p1].Y

	ratio := Div(x-x0, x1-x0)

	return Mul(y0, DecimalUnit-ratio) + Mul(y1, ratio)
}
