package render

import "testing"

func TestMarchingCubes(t *testing.T) {
	max := 0
	for _, tri := range mcTriangleTable {
		if len(tri) > max {
			max = len(tri)
		}
	}
	if max != marchingCubesMaxTriangles {
		t.Error("mismatch marching cubes max triangles")
	}
}
