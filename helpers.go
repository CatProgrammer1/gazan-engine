package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"math"

	_ "image/png"

	_ "image/jpeg"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/qmuntal/gltf"
)

func readIndices(doc *gltf.Document, prim *gltf.Primitive) []uint32 {
	if prim.Indices == nil {
		return nil
	}

	accessor := doc.Accessors[*prim.Indices]
	return readScalarUint32(doc, accessor)
}

func readScalarUint32(doc *gltf.Document, accessor *gltf.Accessor) []uint32 {
	data := getAccessorData(doc, accessor)

	out := make([]uint32, accessor.Count)

	r := bytes.NewReader(data)

	for i := 0; i < int(accessor.Count); i++ {
		var v uint32

		switch accessor.ComponentType {
		case gltf.ComponentUbyte:
			var tmp uint8
			binary.Read(r, binary.LittleEndian, &tmp)
			v = uint32(tmp)

		case gltf.ComponentUshort:
			var tmp uint16
			binary.Read(r, binary.LittleEndian, &tmp)
			v = uint32(tmp)

		case gltf.ComponentUint:
			binary.Read(r, binary.LittleEndian, &v)

		default:
			panic("unsupported index component type")
		}

		out[i] = v
	}

	return out
}

func getImage(document *gltf.Document, textureInfo *gltf.TextureInfo) image.Image {
	texture := document.Textures[textureInfo.Index]

	imgGltf := document.Images[*texture.Source]

	bufferView := document.BufferViews[*imgGltf.BufferView]

	buffer := document.Buffers[bufferView.Buffer]

	img, _, err := image.Decode(bytes.NewReader(buffer.Data[bufferView.ByteOffset : bufferView.ByteOffset+bufferView.ByteLength]))
	handle(err)

	return img
}

func resetScale(m mgl32.Mat4) mgl32.Mat4 {
	// Считаем масштаб по столбцам (X, Y, Z оси)
	sx := float32(math.Sqrt(float64(m[0]*m[0] + m[1]*m[1] + m[2]*m[2])))
	sy := float32(math.Sqrt(float64(m[4]*m[4] + m[5]*m[5] + m[6]*m[6])))
	sz := float32(math.Sqrt(float64(m[8]*m[8] + m[9]*m[9] + m[10]*m[10])))

	if sx == 0 {
		sx = 1
	}
	if sy == 0 {
		sy = 1
	}
	if sz == 0 {
		sz = 1
	}

	return mgl32.Mat4{
		// Первый столбец (ось X)
		m[0] / sx, m[1] / sx, m[2] / sx, m[3],
		// Второй столбец (ось Y)
		m[4] / sy, m[5] / sy, m[6] / sy, m[7],
		// Третий столбец (ось Z)
		m[8] / sz, m[9] / sz, m[10] / sz, m[11],
		// Четвертый столбец (Чистая ПОЗИЦИЯ - оставляем без изменений!)
		m[12], m[13], m[14], m[15],
	}
}

func getImageFromTexture(document *gltf.Document, texture *gltf.Texture) image.Image {
	imgGltf := document.Images[*texture.Source]

	bufferView := document.BufferViews[*imgGltf.BufferView]

	buffer := document.Buffers[bufferView.Buffer]

	img, _, err := image.Decode(bytes.NewReader(buffer.Data[bufferView.ByteOffset : bufferView.ByteOffset+bufferView.ByteLength]))
	handle(err)

	return img
}

func readVec4(doc *gltf.Document, accessor *gltf.Accessor) []mgl32.Vec4 {
	data := getAccessorData(doc, accessor)

	out := make([]mgl32.Vec4, accessor.Count)

	r := bytes.NewReader(data)

	for i := 0; i < int(accessor.Count); i++ {
		var v [4]float32
		binary.Read(r, binary.LittleEndian, &v)
		out[i] = v
	}

	return out
}

func readVec3(doc *gltf.Document, accessor *gltf.Accessor) []mgl32.Vec3 {
	data := getAccessorData(doc, accessor)

	out := make([]mgl32.Vec3, accessor.Count)

	r := bytes.NewReader(data)

	for i := 0; i < int(accessor.Count); i++ {
		var v [3]float32
		binary.Read(r, binary.LittleEndian, &v)
		out[i] = v
	}

	return out
}

func readVec2(doc *gltf.Document, accessor *gltf.Accessor) []mgl32.Vec2 {
	data := getAccessorData(doc, accessor)

	out := make([]mgl32.Vec2, accessor.Count)

	r := bytes.NewReader(data)

	for i := 0; i < int(accessor.Count); i++ {
		var v [2]float32
		binary.Read(r, binary.LittleEndian, &v)
		out[i] = v
	}

	return out
}

func readMat4(doc *gltf.Document, accessor *gltf.Accessor) []float32 {
	data := getAccessorData(doc, accessor)

	out := make([]float32, accessor.Count*16)

	r := bytes.NewReader(data)

	binary.Read(r, binary.LittleEndian, &out)

	return out
}

func lerp(a, b, t float32) float32 {
	return a*(1-t) + b*t
}

func vec3Lerp(a, b mgl32.Vec3, t float32) mgl32.Vec3 {
	return mgl32.Vec3{
		lerp(a[0], b[0], t),
		lerp(a[1], b[1], t),
		lerp(a[2], b[2], t),
	}
}

func getBindMatrix(doc *gltf.Document, skin *gltf.Skin, jointIndex int) mgl32.Mat4 {
	if skin.InverseBindMatrices == nil {
		return mgl32.Ident4()
	}

	accessor := doc.Accessors[*skin.InverseBindMatrices]

	all := readMat4(doc, accessor)

	if len(all) < (jointIndex+1)*16 {
		return mgl32.Ident4()
	}

	start := jointIndex * 16

	var m mgl32.Mat4
	copy(m[:], all[start:start+16])

	return mgl32.Mat4{
		m[0], m[1], m[2], m[3],
		m[4], m[5], m[6], m[7],
		m[8], m[9], m[10], m[11],
		m[12], m[13], m[14], m[15],
	}
}

func getAccessorData(doc *gltf.Document, accessor *gltf.Accessor) []byte {
	if accessor.BufferView == nil {
		panic("accessor has no bufferView")
	}

	view := doc.BufferViews[*accessor.BufferView]
	buffer := doc.Buffers[view.Buffer]

	data := buffer.Data

	start := int(view.ByteOffset + accessor.ByteOffset)

	// Calculate end based on accessor count and element size
	// For matrices (Mat4), we need 16 floats per matrix = 64 bytes
	var elementSize int
	switch accessor.Type {
	case gltf.AccessorMat4:
		elementSize = 64 // 16 floats * 4 bytes
	case gltf.AccessorVec3:
		elementSize = 12 // 3 floats * 4 bytes
	case gltf.AccessorVec4:
		elementSize = 16 // 4 floats * 4 bytes
	case gltf.AccessorScalar:
		elementSize = 4 // 1 float * 4 bytes
	default:
		// Fallback to view.ByteLength
		elementSize = int(view.ByteLength) / int(accessor.Count)
	}

	end := start + int(accessor.Count)*elementSize

	if end > len(data) {
		end = len(data)
	}

	return data[start:end]
}

func isNaN(v float32) bool {
	return math.IsNaN(float64(v))
}

func isInf(v float32) bool {
	return math.IsInf(float64(v), 0)
}

func debugVec3(v [3]float32) {
	fmt.Println(v[0], v[1], v[2])
}
