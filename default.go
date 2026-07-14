package main

import (
	"gl/yks"
	"image"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	defTextureSize = 512
)

var (
	defaultDiffuseTexture,
	defaultNormalTexture,
	defaultMetallicRoughnessTexture *Texture2D

	defaultMaterial *Material

	quadMesh *Mesh

	quadVertices = []Vertex{
		{[3]float32{-1.0, 1.0, 0}, [3]float32{0.0, 0.0, 0}, [2]float32{0, 1}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{-1.0, -1.0, 0}, [3]float32{0.0, 0.0, 0}, [2]float32{0, 0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{1.0, -1.0, 0}, [3]float32{0.0, 0.0, 0}, [2]float32{1, 0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},

		//

		{[3]float32{-1.0, 1.0, 0}, [3]float32{0.0, 0.0, 0}, [2]float32{0, 1}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{1.0, -1.0, 0}, [3]float32{0.0, 0.0, 0}, [2]float32{1, 0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{1.0, 1.0, 0}, [3]float32{0.0, 0.0, 0}, [2]float32{1, 1.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
	}
)

func initDefault() {
	for i := range staticIdentities {
		staticIdentities[i] = mgl32.Ident4()
	}

	for i := 0; i < len(argumentsTemplates); i++ {
		argumentsTemplates[i] = make([]yks.Node, i)
	}

	defaultDiffuseTexture = newTexture(image.Rect(0, 0, 1, 1), []uint8{160, 160, 160, 255}, gl.TEXTURE0)

	defaultNormalTexture = newTexture(image.Rect(0, 0, 1, 1), []uint8{128, 128, 255, 255}, gl.TEXTURE1)

	defaultMetallicRoughnessTexture = newTexture(image.Rect(0, 0, 1, 1), []uint8{255, 255, 255, 255}, gl.TEXTURE2)

	defaultMaterial = newMaterial(defaultDiffuseTexture, defaultNormalTexture, defaultMetallicRoughnessTexture, 1, 1, 12, 0, 0)

	quadMesh = newMesh(quadVertices, nil, nil, gl.STATIC_DRAW,
		Attribute{0, 3, gl.FLOAT, false, vertexStride, 0},
		//Attribute{1, 3, gl.FLOAT, false, vertexStride, uintptr(3 * 4)},
		Attribute{2, 2, gl.FLOAT, false, vertexStride, uintptr(6 * 4)},
	)
}
