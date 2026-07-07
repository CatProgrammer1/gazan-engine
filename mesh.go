package main

import (
	"encoding/binary"
	"fmt"

	"log"
	"os"
	"unsafe"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/qmuntal/gltf"
)

type Attribute struct {
	Index      uint32
	Size       int32
	XType      uint32
	Normalized bool
	Stride     int32
	Pointer    uintptr
}

type Vertex struct {
	Position, Normal [3]float32
	UV               [2]float32
	Joints           [4]uint8
	Weights          [4]float32
	Tangent          [4]float32
}

type SubMesh struct {
	IndexOffset, IndexCount int
	Material                *Material
}

var (
	loadedMeshes = make(chan *Mesh, 25)
)

func newMesh(vertices []Vertex, indices []uint32, submeshes []*SubMesh, usage uint32, atts ...Attribute) *Mesh {
	if len(vertices) == 0 {
		return &Mesh{}
	}

	mesh := &Mesh{
		SubMeshes: submeshes,
	}

	mesh.Vertices = vertices
	mesh.Indices = indices

	var VBO, VAO, EBO uint32
	gl.GenVertexArrays(1, &VAO)
	gl.GenBuffers(1, &VBO)
	gl.GenBuffers(1, &EBO)

	gl.BindVertexArray(VAO)

	size := len(vertices) * int(unsafe.Sizeof(vertices[0]))

	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, size, unsafe.Pointer(&vertices[0]), usage)

	if indices != nil {
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, EBO)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, unsafe.Pointer(&indices[0]), usage)
	}

	for _, attribute := range atts {
		switch attribute.XType {
		case gl.UNSIGNED_BYTE:
			gl.VertexAttribIPointer(attribute.Index, attribute.Size, attribute.XType, attribute.Stride, unsafe.Pointer(attribute.Pointer))
		case gl.FLOAT:
			gl.VertexAttribPointer(attribute.Index, attribute.Size, attribute.XType, attribute.Normalized, attribute.Stride, unsafe.Pointer(attribute.Pointer))
		}
		gl.EnableVertexAttribArray(uint32(attribute.Index))
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	gl.BindVertexArray(0)

	mesh.EBO = EBO
	mesh.VAO = VAO
	mesh.VBO = VBO

	return mesh
}

func newMeshFromFile(path string, usage uint32, parallel bool, atts ...Attribute) *Mesh {
	glbFile, err := os.Open(path)
	handle(err)

	document := gltf.NewDocument()

	gltf.NewDecoder(glbFile).Decode(document)

	glbFile.Close()

	globalIndices := []uint32{}
	globalVertices := []Vertex{}
	subMeshes := []*SubMesh{}

	boneIndexMap, bonesInfo := parseDocumentSkins(document)

	animations := parseDocumentAnimations(document)

	for _, mesh := range document.Meshes {
		for _, primitive := range mesh.Primitives {
			submesh := &SubMesh{}

			var positions []mgl32.Vec3
			var normals []mgl32.Vec3
			var uvs []mgl32.Vec2
			var weights []mgl32.Vec4
			var joints [][4]uint8
			var tangents []mgl32.Vec4

			fmt.Println(primitive.Attributes)

			var materialIndex int
			if primitive.Material != nil {
				materialIndex = *primitive.Material
			} else {
				materialIndex = -1
			}

			var material *gltf.Material
			if materialIndex != -1 {
				material = document.Materials[materialIndex]
			}

			var diffuseTexture2D, normalTexture2D, roughnessTexture2D *Texture2D

			var pbr *gltf.PBRMetallicRoughness
			if material != nil {
				pbr = material.PBRMetallicRoughness
			}

			var roughnessFactor, metallicFactor float32
			if pbr != nil {
				roughnessFactor = float32(pbr.RoughnessFactorOrDefault())
				metallicFactor = float32(pbr.MetallicFactorOrDefault())
			}

			if pbr != nil && pbr.BaseColorTexture != nil {
				diffuseImg := getImage(document, pbr.BaseColorTexture)

				if parallel {
					safeParallellGlChannelInput <- func() []any {
						tex := newTextureFromImage(diffuseImg, gl.TEXTURE0)
						return []any{tex}
					}

					diffuseTexture2D = (<-safeParallellGlChannelOutput)[0].(*Texture2D)
				} else {
					diffuseTexture2D = newTextureFromImage(diffuseImg, gl.TEXTURE0)
				}
			}

			if material != nil && material.NormalTexture != nil {
				texture := document.Textures[*material.NormalTexture.Index]

				normalImg := getImageFromTexture(document, texture)

				if parallel {
					safeParallellGlChannelInput <- func() []any {
						tex := newTextureFromImage(normalImg, gl.TEXTURE1)
						return []any{tex}
					}

					normalTexture2D = (<-safeParallellGlChannelOutput)[0].(*Texture2D)
				} else {
					normalTexture2D = newTextureFromImage(normalImg, gl.TEXTURE1)
				}
			}

			if pbr != nil && pbr.MetallicRoughnessTexture != nil {
				texture := document.Textures[pbr.MetallicRoughnessTexture.Index]

				roughnessImg := getImageFromTexture(document, texture)

				if parallel {
					safeParallellGlChannelInput <- func() []any {
						tex := newTextureFromImage(roughnessImg, gl.TEXTURE2)
						return []any{tex}
					}

					roughnessTexture2D = (<-safeParallellGlChannelOutput)[0].(*Texture2D)
				} else {
					roughnessTexture2D = newTextureFromImage(roughnessImg, gl.TEXTURE2)
				}
			}
			// POSITION
			if attr, ok := primitive.Attributes["POSITION"]; ok {
				accessor := document.Accessors[attr]
				positions = readVec3(document, accessor)
			}

			if attr, ok := primitive.Attributes["NORMAL"]; ok {
				accessor := document.Accessors[attr]
				normals = readVec3(document, accessor)
			}

			if attr, ok := primitive.Attributes["TEXCOORD_0"]; ok {
				accessor := document.Accessors[attr]
				uvs = readVec2(document, accessor)
			}

			if attr, ok := primitive.Attributes["TANGENT"]; ok {
				accessor := document.Accessors[attr]

				tangents = readVec4(document, accessor)
			}

			if attr, ok := primitive.Attributes["WEIGHTS_0"]; ok {
				accessor := document.Accessors[attr]

				weightsBytes := getAccessorData(document, accessor)

				readyWeights := make([]mgl32.Vec4, accessor.Count)
				binary.Decode(weightsBytes, binary.LittleEndian, &readyWeights)

				weights = readyWeights
			}

			if attr, ok := primitive.Attributes["JOINTS_0"]; ok {
				accessor := document.Accessors[attr]

				jointsBytes := getAccessorData(document, accessor)

				readyJoints := make([][4]uint8, accessor.Count)
				binary.Decode(jointsBytes, binary.LittleEndian, &readyJoints)

				joints = readyJoints
			}

			meshVertices := make([]Vertex, len(positions))

			indices := readIndices(document, primitive)

			baseVertex := uint32(len(globalVertices))

			if diffuseTexture2D != nil {
				if parallel {
					safeParallellGlChannelInput <- func() []any {
						return []any{newMaterial(diffuseTexture2D, normalTexture2D, roughnessTexture2D, roughnessFactor, metallicFactor, 14, 0, 0)}
					}

					submesh.Material = (<-safeParallellGlChannelOutput)[0].(*Material)
				} else {
					submesh.Material = newMaterial(diffuseTexture2D, normalTexture2D, roughnessTexture2D, roughnessFactor, metallicFactor, 14, 0, 0)
				}
			}
			submesh.IndexOffset = len(globalIndices)
			submesh.IndexCount = len(indices)

			for _, i := range indices {
				uv := mgl32.Vec2{uvs[i].X(), 1.0 - uvs[i].Y()}

				vertexWeight := [4]float32{1, 0, 0, 0}
				vertexJoints := [4]uint8{1, 1, 1, 1}

				if len(joints) > 0 && len(weights) > 0 {
					vertexWeight = weights[i]
					vertexJoints = joints[i]
				}

				if len(tangents) < int(i) {
					throwf("::Mesh - '%s' has too few or doesn't have any tangents.\n", path)
				}

				if len(normals) < int(i) {
					throwf("::Mesh - '%s' has too few or doesn't have any normals.\n", path)
				}

				meshVertices[i] = Vertex{
					Position: positions[i],
					Normal:   normals[i],
					UV:       uv,
					Weights:  vertexWeight,
					Joints:   vertexJoints,
					Tangent:  tangents[i],
				}
				globalIndices = append(globalIndices, baseVertex+uint32(i))
			}

			globalVertices = append(globalVertices, meshVertices...)

			subMeshes = append(subMeshes, submesh)
		}
	}

	if parallel {
		safeParallellGlChannelInput <- func() []any {
			newMesh := newMesh(globalVertices, globalIndices, subMeshes, usage, atts...)
			newMesh.Animations = animations
			newMesh.bonesInfo = bonesInfo
			newMesh.boneIndexMap = boneIndexMap
			newMesh.Name = path

			return []any{newMesh}
		}

		mesh := (<-safeParallellGlChannelOutput)[0].(*Mesh)

		return mesh
	}

	newMesh := newMesh(globalVertices, globalIndices, subMeshes, usage, atts...)
	newMesh.Animations = animations
	newMesh.boneIndexMap = boneIndexMap
	newMesh.bonesInfo = bonesInfo
	newMesh.Name = path

	return newMesh
}

func newMeshWithBuffers(VBO, EBO uint32, submeshes []*SubMesh, atts ...Attribute) *Mesh {
	Mesh := &Mesh{
		SubMeshes: submeshes,
	}

	var VAO uint32
	gl.GenVertexArrays(1, &VAO)

	gl.BindVertexArray(VAO)

	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)

	if EBO != 0 {
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, EBO)
	}

	for i, attribute := range atts {
		gl.VertexAttribPointer(attribute.Index, attribute.Size, attribute.XType, attribute.Normalized, attribute.Stride, unsafe.Pointer(attribute.Pointer))
		gl.EnableVertexAttribArray(uint32(i))
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	gl.BindVertexArray(0)

	Mesh.EBO = EBO
	Mesh.VAO = VAO
	Mesh.VBO = VBO

	return Mesh
}

// Changing any fields of an instance of this structure will lead to all of the mesh objects with pointer to this mesh to be changed
// Change meshObject's fields, if you don't need global changes
type Mesh struct {
	Name string

	SubMeshes     []*SubMesh
	Vertices      []Vertex
	Indices       []uint32
	Animations    []*Animation
	boneIndexMap  map[string]int
	bonesInfo     []*BoneInfo
	VAO, VBO, EBO uint32
}

func (mesh Mesh) DrawElements(shaderProgram ShaderProgram, mode uint32, xtype uint32) {
	if len(mesh.Vertices) == 0 {
		return
	}
	if mesh.Indices == nil {
		log.Println("error no indices")
		return
	}

	var xtypeSize int32
	switch xtype {
	case gl.UNSIGNED_INT:
		xtypeSize = 4
	case gl.UNSIGNED_SHORT:
		xtypeSize = 2
	case gl.UNSIGNED_BYTE:
		xtypeSize = 1
	}

	gl.BindVertexArray(mesh.VAO)
	if mesh.SubMeshes != nil {
		for _, submesh := range mesh.SubMeshes {
			if submesh.Material != nil {
				submesh.Material.Use(shaderProgram, "material")
			} else if defaultMaterial != nil {
				defaultMaterial.Use(shaderProgram, "material")
			}
			gl.DrawElements(mode, int32(submesh.IndexCount), xtype, unsafe.Pointer(uintptr(submesh.IndexOffset*int(xtypeSize))))
		}
	}else {
		gl.DrawElements(mode, int32(len(mesh.Indices)), xtype, nil)
	}
	gl.BindVertexArray(0)
}

func (mesh Mesh) DrawArrays(mode uint32, first, count int32) {
	if len(mesh.Vertices) == 0 {
		return
	}
	if mesh.Vertices == nil && mesh.VAO == 0 {
		log.Println("error wth")
		return
	}
	gl.BindVertexArray(mesh.VAO)

	gl.DrawArrays(mode, first, count)

	gl.BindVertexArray(0)
}

func (mesh Mesh) Delete() {
	if len(mesh.Vertices) == 0 {
		return
	}
	gl.DeleteVertexArrays(1, &mesh.VAO)
	gl.DeleteBuffers(1, &mesh.VBO)
	gl.DeleteBuffers(1, &mesh.EBO)
}

func (mesh Mesh) CloneBoneInfo() ([]*BoneInfo, map[string]int) {
	newBoneIndexMap := make(map[string]int)
	newBonesInfo := make([]*BoneInfo, len(mesh.bonesInfo))

	for i, boneInfo := range mesh.bonesInfo {
		newBoneInfo := new(BoneInfo)

		newBoneInfo.ID = boneInfo.ID
		newBoneInfo.FinalTransformation = boneInfo.FinalTransformation
		newBoneInfo.OffsetMatrix = boneInfo.OffsetMatrix

		newBonesInfo[i] = newBoneInfo
	}

	for k, boneIndex := range mesh.boneIndexMap {
		newBoneIndexMap[k] = boneIndex
		newBonesInfo[boneIndex] = mesh.bonesInfo[boneIndex]
	}

	return newBonesInfo, newBoneIndexMap
}
