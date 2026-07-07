package main

import (
	"github.com/go-gl/mathgl/mgl32"
)

func testmath() {
	trans := mgl32.Ident4()
	trans = mgl32.HomogRotate3D(mgl32.DegToRad(90), mgl32.Vec3{0, 0, 1})

	trans = trans.Mul4(mgl32.Scale3D(.5, .5, .5))
}
