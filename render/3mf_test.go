package render_test

import (
	"testing"

	"github.com/drummonds/sdf/form3"
	"github.com/drummonds/sdf/render"
	"gonum.org/v1/gonum/spatial/r3"
)

func Test3MF(t *testing.T) {
	const path = "box.3mf"
	box, _ := form3.Box(r3.Vec{X: 1, Y: 1, Z: 1}, .1)
	err := render.Create3MF(path, render.NewOctreeRenderer(box, 10))
	if err != nil {
		t.Error(err)
	}
}
