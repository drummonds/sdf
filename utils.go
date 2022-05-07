package sdf

import (
	"errors"
	"math"

	"github.com/drummonds/sdf/internal/d2"
	"github.com/drummonds/sdf/internal/d3"
	"gonum.org/v1/gonum/spatial/r2"
	"gonum.org/v1/gonum/spatial/r3"
)

const (
	pi        = math.Pi
	tau       = 2 * pi
	sqrtHalf  = 0.7071067811865476
	tolerance = 1e-9
)

const (
	// epsilon is the machine epsilon. For IEEE this is 2^{-53}.
	dlamchE = 0x1p-53
	// dlamchB is the radix of the machine (the base of the number system).
	dlamchB = 2
	// dlamchP is base * eps.
	dlamchP = dlamchB * dlamchE
	// dlamchS is the "safe minimum", that is, the lowest number such that
	// 1/dlamchS does not overflow, or also the smallest normal number.
	// For IEEE this is 2^{-1022}.
	dlamchS = 0x1p-1022
	epsilon = 1e-12
)

// R2ToI temporary home for this function.
//
// Deprecated: R2ToI is deprecated.
func R2ToI(a r2.Vec) V2i { // Deprecated: R2ToI is deprecated.
	return V2i{int(a.X), int(a.Y)}
}

// R3ToI temporary home for this function.
//
// Deprecated: R3ToI is deprecated.
func R3ToI(a r3.Vec) V3i {
	return V3i{int(a.X), int(a.Y), int(a.Z)}
}

// R3FromI temporary home for this function.
//
//Deprecated: R3FromI is deprecated.
func R3FromI(a V3i) r3.Vec {
	return d3.NewV3(float64(a[0]), float64(a[1]), float64(a[2]))
}

// R2FromI temporary home for this function.
//
// Deprecated: R2FromI is deprecated.
func R2FromI(a V2i) r2.Vec {
	return d2.NewV2(float64(a[0]), float64(a[1]))
}

// clamp x between a and b, assume a <= b
func clamp(x, a, b float64) float64 {
	if x < a {
		return a
	}
	if x > b {
		return b
	}
	return x
}

// mix does a linear interpolation from x to y, a = [0,1]
func mix(x, y, a float64) float64 {
	return x + (a * (y - x))
}

// sign returns the sign of x
func sign(x float64) float64 {
	if x < 0 {
		return -1
	}
	if x > 0 {
		return 1
	}
	return 0
}

// sawTooth generates a sawtooth function. Returns [-period/2, period/2)
func sawTooth(x, period float64) float64 {
	x += period / 2
	t := x / period
	return period*(t-math.Floor(t)) - period/2
}

// MinRound returns a minimum function that uses a quarter-circle to join the two objects smoothly.
func MinRound(k float64) MinFunc {
	return func(a, b float64) float64 {
		u := d2.MaxElem(d2.NewV2(k-a, k-b), d2.NewV2(0, 0))
		return math.Max(k, math.Min(a, b)) - r2.Norm(u)
	}
}

// MinChamfer returns a minimum function that makes a 45-degree chamfered edge (the diagonal of a square of size <r>).
// TODO: why the holes in the rendering?
func MinChamfer(k float64) MinFunc {
	return func(a, b float64) float64 {
		return math.Min(math.Min(a, b), (a-k+b)*sqrtHalf)
	}
}

// MinExp returns a minimum function with exponential smoothing (k = 32).
func MinExp(k float64) MinFunc {
	return func(a, b float64) float64 {
		return -math.Log(math.Exp(-k*a)+math.Exp(-k*b)) / k
	}
}

// MinPow returns  a minimum function (k = 8).
// TODO - weird results, is this correct?
func MinPow(k float64) MinFunc {
	return func(a, b float64) float64 {
		a = math.Pow(a, k)
		b = math.Pow(b, k)
		return math.Pow((a*b)/(a+b), 1/k)
	}
}

func poly(a, b, k float64) float64 {
	h := clamp(0.5+0.5*(b-a)/k, 0.0, 1.0)
	return mix(b, a, h) - k*h*(1.0-h)
}

// MinPoly returns a minimum function (Try k = 0.1, a bigger k gives a bigger fillet).
func MinPoly(k float64) MinFunc {
	return func(a, b float64) float64 {
		return poly(a, b, k)
	}
}

// MaxFunc is a maximum function for SDF blending.
type MaxFunc func(a, b float64) float64

// MaxPoly returns a maximum function (Try k = 0.1, a bigger k gives a bigger fillet).
func MaxPoly(k float64) MaxFunc {
	return func(a, b float64) float64 {
		return -poly(-a, -b, k)
	}
}

// ExtrudeFunc maps r3.Vec to V2 - the point used to evaluate the SDF2.
type ExtrudeFunc func(p r3.Vec) r2.Vec

// NormalExtrude returns an extrusion function.
func NormalExtrude(p r3.Vec) r2.Vec {
	return d2.NewV2(p.X, p.Y)
}

// TwistExtrude returns an extrusion function that twists with z.
func TwistExtrude(height, twist float64) ExtrudeFunc {
	k := twist / height
	return func(p r3.Vec) r2.Vec {
		m := Rotate(p.Z * k)
		return m.MulPosition(d2.NewV2(p.X, p.Y))
	}
}

// ScaleExtrude returns an extrusion functions that scales with z.
func ScaleExtrude(height float64, scale r2.Vec) ExtrudeFunc {
	inv := d2.NewV2(1/scale.X, 1/scale.Y)
	// TODO verify
	m := d2.DivElem(inv.Sub(d2.NewV2(1, 1)), d2.Elem(height)) // slope
	b := r2.Add(d2.DivElem(inv, d2.Elem(2)), d2.Elem(0.5))
	// b := inv.DivScalar(2).AddScalar(0.5)     // intercept
	return func(p r3.Vec) r2.Vec {
		return d2.MulElem(d2.NewV2(p.X, p.Y), r2.Scale(p.Z, m).Add(b))
	}
}

// ScaleTwistExtrude returns an extrusion function that scales and twists with z.
func ScaleTwistExtrude(height, twist float64, scale r2.Vec) ExtrudeFunc {
	k := twist / height
	inv := d2.NewV2(1/scale.X, 1/scale.Y)
	m := r2.Sub(inv, d2.DivElem(d2.NewV2(1, 1), d2.Elem(height))) // slope
	// m := inv.Sub(d2.NewV2(1, 1)).DivScalar(height) // slope
	b := r2.Add(d2.DivElem(inv, d2.Elem(2)), d2.Elem(0.5))
	// b := inv.DivScalar(2).AddScalar(0.5) // intercept
	return func(p r3.Vec) r2.Vec {
		// Scale and then Twist
		// pnew := d2.NewV2(p.X, p.Y).Mul(m.MulScalar(p.Z).Add(b)) // Scale
		pnew := d2.MulElem(d2.NewV2(p.X, p.Y), r2.Add(r2.Scale(p.Z, m), b))
		return Rotate(p.Z * k).MulPosition(pnew) // Twist

		// Twist and then scale
		//pnew := Rotate(p.Z * k).MulPosition(d2.NewV2(p.X, p.Y))
		//return pnew.Mul(m.MulScalar(p.Z).Add(b))
	}
}

// Raycasting

func sigmoidScaled(x float64) float64 {
	return 2/(1+math.Exp(-x)) - 1
}

// raycast3 collides a ray (with an origin point from and a direction dir) with an SDF3.
// sigmoid is useful for fixing bad distance functions (those that do not accurately represent the distance to the
// closest surface, but will probably imply more evaluations)
// stepScale controls precision (less stepSize, more precision, but more SDF evaluations): use 1 if SDF indicates
// distance to the closest surface.
// It returns the collision point, how many normalized distances to reach it (t), and the number of steps performed
// If no surface is found (in maxDist and maxSteps), t is < 0
func raycast3(s SDF3, from, dir r3.Vec, scaleAndSigmoid, stepScale, epsilon, maxDist float64, maxSteps int) (collision r3.Vec, t float64, steps int) {
	t = 0
	dirN := r3.Unit(dir)
	pos := from
	for {
		val := math.Abs(s.Evaluate(pos))
		//log.Print("Raycast step #", steps, " at ", pos, " with value ", val, "\n")
		if val < epsilon {
			collision = pos // Success
			break
		}
		steps++
		if steps == maxSteps {
			t = -1 // Failure
			break
		}
		if scaleAndSigmoid > 0 {
			val = sigmoidScaled(val * 10)
		}
		delta := val * stepScale
		t += delta
		pos = r3.Add(pos, r3.Scale(delta, dirN))
		if t < 0 || t > maxDist {
			t = -1 // Failure
			break
		}
	}
	//log.Println("Raycast did", steps, "steps")
	return
}

// raycast2 see Raycast3. NOTE: implementation using Raycast3 (inefficient?)
func raycast2(s SDF2, from, dir r2.Vec, scaleAndSigmoid, stepScale, epsilon, maxDist float64, maxSteps int) (r2.Vec, float64, int) {
	collision, t, steps := raycast3(Extrude3D(s, 1),
		d3.NewV3(from.X, from.Y, 0),
		d3.NewV3(dir.X, dir.Y, 0),
		scaleAndSigmoid, stepScale, epsilon, maxDist, maxSteps)
	return d2.NewV2(collision.X, collision.Y), t, steps
}

// Normals

// normal3 returns the normal of an SDF3 at a point (doesn't need to be on the surface).
// Computed by sampling it several times inside a box of side 2*eps centered on p.
func normal3(s SDF3, p r3.Vec, eps float64) r3.Vec {
	return r3.Unit(r3.Vec{
		X: s.Evaluate(p.Add(r3.Vec{X: eps})) - s.Evaluate(p.Add(r3.Vec{X: -eps})),
		Y: s.Evaluate(p.Add(r3.Vec{Y: eps})) - s.Evaluate(p.Add(r3.Vec{Y: -eps})),
		Z: s.Evaluate(p.Add(r3.Vec{Z: eps})) - s.Evaluate(p.Add(r3.Vec{Z: -eps})),
	})
}

// normal2 returns the normal of an SDF3 at a point (doesn't need to be on the surface).
// Computed by sampling it several times inside a box of side 2*eps centered on p.
func normal2(s SDF2, p r2.Vec, eps float64) r2.Vec {
	return r2.Unit(r2.Vec{
		X: s.Evaluate(p.Add(r2.Vec{X: eps})) - s.Evaluate(p.Add(r2.Vec{X: -eps})),
		Y: s.Evaluate(p.Add(r2.Vec{Y: eps})) - s.Evaluate(p.Add(r2.Vec{Y: -eps})),
	})
}

// Floating Point Comparisons
// See: http://floating-point-gui.de/errors/NearlyEqualsTest.java

const minNormal = 2.2250738585072014e-308 // 2**-1022

// MulVertices multiples a set of V2 vertices by a rotate/translate matrix.
func mulVertices2(v d2.Set, a m33) {
	for i := range v {
		v[i] = a.MulPosition(v[i])
	}
}

// MulVertices multiples a set of r3.Vec vertices by a rotate/translate matrix.
func mulVertices3(v d3.Set, a m44) {
	for i := range v {
		v[i] = a.MulPosition(v[i])
	}
}

// map2 maps a 2d region to integer grid coordinates.
type map2 struct {
	bb    d2.Box // bounding box
	grid  V2i    // integral dimension
	delta r2.Vec
	flipy bool // flip the y-axis
}

// newMap2 returns a 2d region to grid coordinates map.
func newMap2(bb d2.Box, grid V2i, flipy bool) (*map2, error) {
	// sanity check the bounding box
	bbSize := bb.Size()
	if bbSize.X <= 0 || bbSize.Y <= 0 {
		return nil, errors.New("bad bounding box")
	}
	// sanity check the integer dimensions
	if grid[0] <= 0 || grid[1] <= 0 {
		return nil, errors.New("bad grid dimensions")
	}
	m := map2{}
	m.bb = bb
	m.grid = grid
	m.flipy = flipy
	m.delta = d2.DivElem(bbSize, R2FromI(grid))
	return &m, nil
}

// ToVec converts grid integer coordinates to 2d region float coordinates.
func (m *map2) ToV2(p V2i) r2.Vec {
	ofs := d2.MulElem(r2.Add(R2FromI(p), d2.Elem(0.5)), m.delta)
	// ofs := p.ToV2().AddScalar(0.5).Mul(m.delta)
	var origin r2.Vec
	if m.flipy {
		origin = m.bb.TopLeft()
		ofs.Y = -ofs.Y
	} else {
		origin = m.bb.BottomLeft()
	}
	return origin.Add(ofs)
}

// ToV2i converts 2d region float coordinates to grid integer coordinates.
func (m *map2) ToV2i(p r2.Vec) V2i {
	var v r2.Vec
	if m.flipy {
		v = p.Sub(m.bb.TopLeft())
		v.Y = -v.Y
	} else {
		v = p.Sub(m.bb.BottomLeft())
	}
	return R2ToI(d2.DivElem(v, m.delta)) // v.Div(m.delta).ToV2i()
}

func sdfBox3d(p, s r3.Vec) float64 {
	d := r3.Sub(d3.AbsElem(p), s)
	if d.X > 0 && d.Y > 0 && d.Z > 0 {
		return r3.Norm(d)
	}
	if d.X > 0 && d.Y > 0 {
		return math.Hypot(d.X, d.Y) // V2{d.X, d.Y}.Length()
	}
	if d.X > 0 && d.Z > 0 {
		return math.Hypot(d.X, d.Z) // V2{d.X, d.Z}.Length()
	}
	if d.Y > 0 && d.Z > 0 {
		return math.Hypot(d.Y, d.Z) //V2{d.Y, d.Z}.Length()
	}
	if d.X > 0 {
		return d.X
	}
	if d.Y > 0 {
		return d.Y
	}
	if d.Z > 0 {
		return d.Z
	}
	return d3.Max(d)
}
