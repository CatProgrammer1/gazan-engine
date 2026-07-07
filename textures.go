package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"regexp"
	"unsafe"

	"github.com/disintegration/imaging"
	"github.com/go-gl/gl/v4.3-core/gl"
)

var (
	formatRegex = regexp.MustCompile(`\.([^.]+)$`)
)

func newTextureFromFile(path string, textureUnit uint32) *Texture2D {
	file, err := os.Open(path)
	handle(err)
	defer file.Close()

	img, _, err := image.Decode(file)
	handle(err)

	nrgba := imaging.FlipV(img)

	return newTexture(img.Bounds(), nrgba.Pix, textureUnit)
}

func newTextureFromImage(img image.Image, textureUnit uint32) *Texture2D {
	nrgba := imaging.FlipV(img)

	return newTexture(img.Bounds(), nrgba.Pix, textureUnit)
}

func newTexture(bounds image.Rectangle, pix []uint8, textureUnit uint32) *Texture2D {
	texture2D := &Texture2D{
		pix:         pix,
		Bounds:      bounds,
		textureUnit: textureUnit,
	}

	gl.GenTextures(1, &texture2D.texture)

	gl.BindTexture(gl.TEXTURE_2D, texture2D.texture)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(bounds.Dx()),
		int32(bounds.Dy()),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		unsafe.Pointer(&pix[0]),
	)
	gl.GenerateMipmap(gl.TEXTURE_2D)
	gl.BindTexture(gl.TEXTURE_2D, 0)

	return texture2D
}

type Texture2D struct {
	texture     uint32
	pix         []uint8
	textureUnit uint32
	Bounds      image.Rectangle
}

func (tex2d Texture2D) SetSampleUnit(shaderProgram ShaderProgram, uniform string, index int32) {
	gl.Uniform1i(shaderProgram.GetUniformLocation(uniform), index)
}

func (tex2d Texture2D) SetSampleUnitLocation(shaderProgram ShaderProgram, location int32, index int32) {
	gl.Uniform1i(location, index)
}

func (tex2d Texture2D) Bind() {
	gl.ActiveTexture(tex2d.textureUnit)
	gl.BindTexture(gl.TEXTURE_2D, tex2d.texture)
}
