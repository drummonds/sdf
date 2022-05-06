package form3

import (
	"runtime/debug"

	"github.com/drummonds/sdf"
	"github.com/drummonds/sdf/form3/must3"
)

// Screw returns a screw SDF3.
// - length of screw
// - thread taper angle (radians)
// - pitch thread to thread distance
// - number of thread starts (< 0 for left hand threads)
func Screw(thread sdf.SDF2, length, taper, pitch float64, starts int) (s sdf.SDF3, err error) {
	defer func() {
		if a := recover(); a != nil {
			err = &shapeErr{
				panicObj: a,
				stack:    string(debug.Stack()),
			}
		}
	}()
	return must3.Screw(thread, length, taper, pitch, starts), err
}
