package main

import (
	"image/draw"
	"os"
	"unsafe"

	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/go-gl/gl/v4.3-core/gl"
)

func newCubeMapFromFile(path string) *CubeMap {
	file, err := os.Open(path)
	handle(err)
	defer file.Close()

	img, _, err := image.Decode(file)
	handle(err)

	b := img.Bounds()
	nrgba := image.NewRGBA(b)

	draw.Draw(nrgba, b, img, b.Min, draw.Src)

	return newCubeMap(img.Bounds(), gl.RGBA, gl.UNSIGNED_BYTE, gl.LINEAR, [][]uint8{nrgba.Pix, nrgba.Pix, nrgba.Pix, nrgba.Pix, nrgba.Pix, nrgba.Pix})
}

func newCubeMap(bounds image.Rectangle, format int32, xtype uint32, param int32, pixs [][]uint8) *CubeMap {
	cubeMap := &CubeMap{
		Bounds: bounds,
		unit:   9,
	}

	gl.GenTextures(1, &cubeMap.texture)

	gl.BindTexture(gl.TEXTURE_CUBE_MAP, cubeMap.texture)

	for i := uint32(0); i < 6; i++ {
		if pixs == nil {
			gl.TexImage2D(
				gl.TEXTURE_CUBE_MAP_POSITIVE_X+i,
				0,
				format,
				int32(bounds.Max.X),
				int32(bounds.Max.Y),
				0,
				uint32(format),
				xtype,
				nil,
			)
			continue
		}
		gl.TexImage2D(
			gl.TEXTURE_CUBE_MAP_POSITIVE_X+i,
			0,
			format,
			int32(bounds.Max.X),
			int32(bounds.Max.Y),
			0,
			uint32(format),
			xtype,
			unsafe.Pointer(&pixs[i][0]),
		)
	}

	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MIN_FILTER, param)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MAG_FILTER, param)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)

	gl.BindTexture(gl.TEXTURE_CUBE_MAP, 0)

	return cubeMap
}

type CubeMap struct {
	texture uint32
	unit    int32
	Bounds  image.Rectangle
}

func (tex2d *CubeMap) Bind(location int32) {
	gl.ActiveTexture(gl.TEXTURE0 + uint32(tex2d.unit))
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, tex2d.texture)

	gl.Uniform1i(location, tex2d.unit)
}
