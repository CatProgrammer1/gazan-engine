package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"gl/yks"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/qmuntal/gltf"
)

type Sampler struct {
	Time          []float32
	Interpolation gltf.Interpolation
	Value3        []mgl32.Vec3
	Value4        []mgl32.Vec4
	Path          string
}

type Channel struct {
	Node    *gltf.Node
	Path    gltf.TRSProperty
	Sampler int
}

type TRS struct {
	T mgl32.Vec3
	R mgl32.Vec4
	S mgl32.Vec3
}

type Animation struct {
	Name string

	Document   *gltf.Document
	Mesh       *Mesh
	MeshObject *MeshObject
	Samplers   []Sampler
	Channels   []Channel
	Transforms []mgl32.Mat4
	TimeMarker float32
	IsPlaying  bool
	Looped     bool
	LastTime   float32

	ScriptAnimation *yks.StructObject
}

func (animation *Animation) Play(currentTime float32) {
	if animation.MeshObject == nil {
		Throw("Animation must be have a meshObject")
	}

	animation.IsPlaying = true
	animation.TimeMarker = 0
	animation.LastTime = currentTime
}

func (animation *Animation) Update(currentTime float32) {
	if animation.MeshObject == nil {
		Throw("Animation must be have a meshObject")
	}

	document := animation.Document

	if animation.IsPlaying {
		lastTime := animation.LastTime
		t := currentTime - lastTime

		animation.TimeMarker = t

		scene := document.Scenes[*document.Scene]

		for _, channel := range animation.Channels {
			sampler := animation.Samplers[channel.Sampler]
			node := channel.Node

			lastTime := sampler.Time[len(sampler.Time)-1]
			if t >= lastTime {
				if animation.Looped {
					animation.Play(currentTime)
					return
				} else {
					animation.IsPlaying = false
					return
				}
			}

			for i, t0 := range sampler.Time {
				endI := i + 1
				if t >= t0 && endI < len(sampler.Time) && t <= sampler.Time[endI] {
					t1 := sampler.Time[endI]
					alpha := (t - t0) / (t1 - t0)

					switch sampler.Path {
					case "rotation":
						v4_0, v4_1 := sampler.Value4[i], sampler.Value4[endI]

						q := mgl32.QuatSlerp(
							mgl32.Quat{
								W: v4_0[3],
								V: v4_0.Vec3(),
							},
							mgl32.Quat{
								W: v4_1[3],
								V: v4_1.Vec3(),
							},
							alpha,
						)

						qV := q.V
						qW := q.W

						node.Rotation = [4]float64{float64(qV[0]), float64(qV[1]), float64(qV[2]), float64(qW)}
					case "translation":
						v0, v1 := sampler.Value3[i], sampler.Value3[endI]

						v := vec3Lerp(v0, v1, alpha)

						node.Translation = [3]float64{float64(v[0]), float64(v[1]), float64(v[2])}
					case "scale":
						v0, v1 := sampler.Value3[i], sampler.Value3[endI]

						v := vec3Lerp(v0, v1, alpha)

						node.Scale = [3]float64{float64(v[0]), float64(v[1]), float64(v[2])}
					}
					break
				}
			}
		}

		for _, rootNodeIndex := range scene.Nodes {
			rootNode := document.Nodes[rootNodeIndex]

			animation.Transforms = animation.getNodeTransforms(rootNode)
		}
	}
}

func (animation *Animation) GetBone(name string) *BoneInfo {
	if animation.MeshObject == nil {
		Throw("Animation must be have a meshObject")
	}
	meshObject := animation.MeshObject

	index, ok := meshObject.boneIndexMap[name]
	if !ok {
		return nil
	}

	boneInfo := meshObject.bonesInfo[index]

	return boneInfo
}

func (animation *Animation) getNodeTransforms(node *gltf.Node) []mgl32.Mat4 {
	bonesInfo := animation.MeshObject.bonesInfo
	transforms := make([]mgl32.Mat4, len(bonesInfo))

	// Initialize all transforms to identity
	for i := range transforms {
		transforms[i] = mgl32.Ident4()
	}

	animation.readNodeHierarchy(node, mgl32.Ident4())

	for i, boneInfo := range bonesInfo {
		transforms[i] = boneInfo.FinalTransformation
	}

	return transforms
}

func getNodeTransformations(node *gltf.Node) TRS {
	rotationVec4 := mgl32.Vec4{float32(node.Rotation[0]), float32(node.Rotation[1]), float32(node.Rotation[2]), float32(node.Rotation[3])}

	return TRS{mgl32.Vec3{float32(node.Translation[0]), float32(node.Translation[1]), float32(node.Translation[2])},
		rotationVec4,
		mgl32.Vec3{float32(node.Scale[0]), float32(node.Scale[1]), float32(node.Scale[2])}}
}

func trsMat4(trs TRS) mgl32.Mat4 {
	ident := mgl32.Ident4()

	q := mgl32.Quat{
		W: trs.R[3],
		V: trs.R.Vec3(),
	}.Normalize()

	ident = ident.Mul4(mgl32.Translate3D(trs.T[0], trs.T[1], trs.T[2]))
	ident = ident.Mul4(q.Mat4())
	ident = ident.Mul4(mgl32.Scale3D(trs.S[0], trs.S[1], trs.S[2]))

	return ident
}

func (animation *Animation) readNodeHierarchy(node *gltf.Node, parentTransform mgl32.Mat4) {
	document := animation.Document
	bonesInfo := animation.MeshObject.bonesInfo
	boneIndexMap := animation.MeshObject.boneIndexMap

	trs := getNodeTransformations(node)
	nodeTransformation := trsMat4(trs)

	globalTransform := parentTransform.Mul4(nodeTransformation)

	if boneIndex, ok := boneIndexMap[node.Name]; ok {
		boneInfo := bonesInfo[boneIndex]
		boneInfo.GlobalTransformation = globalTransform
		boneInfo.FinalTransformation = globalTransform.Mul4(boneInfo.OffsetMatrix)
	}

	for _, childIndex := range node.Children {
		child := document.Nodes[childIndex]

		animation.readNodeHierarchy(child, globalTransform)
	}
}

func parse_node(document *gltf.Document, node *gltf.Node, i int) {
	tabs := strings.Repeat("   ", i)
	fmt.Printf(tabs+"::Node name - '%s', Mesh - %v;\n", node.Name, node.Mesh)
	fmt.Println(tabs+"::TRS -", node.TranslationOrDefault(), node.RotationOrDefault(), node.ScaleOrDefault())

	for _, childIndex := range node.Children {
		child := document.Nodes[childIndex]

		parse_node(document, child, i+1)
	}
}

func parseDocumentAnimations(document *gltf.Document) []*Animation {
	animations := []*Animation{}

	samplerIndexToPath := make(map[int]string)

	for _, docAnimation := range document.Animations {
		animation := &Animation{
			Name: docAnimation.Name,

			Document: document,
			Samplers: make([]Sampler, len(docAnimation.Samplers)),
			Channels: make([]Channel, len(docAnimation.Channels)),
		}
		for i, docChannel := range docAnimation.Channels {
			node := document.Nodes[*docChannel.Target.Node]
			path := docChannel.Target.Path

			samplerIndexToPath[docChannel.Sampler] = path.String()

			animation.Channels[i] = Channel{
				Node:    node,
				Path:    path,
				Sampler: docChannel.Sampler,
			}
		}

		for i, docSampler := range docAnimation.Samplers {
			outputAccessor := document.Accessors[docSampler.Output]
			inputAccessor := document.Accessors[docSampler.Input]

			rOutput := bytes.NewReader(getAccessorData(document, outputAccessor))
			rInput := bytes.NewReader(getAccessorData(document, inputAccessor))

			sampler := Sampler{}

			path := samplerIndexToPath[i]

			switch path {
			case "rotation":
				vals := make([]mgl32.Vec4, outputAccessor.Count)
				binary.Read(rOutput, binary.LittleEndian, &vals)

				sampler.Value4 = vals
			default:
				vals := make([]mgl32.Vec3, outputAccessor.Count)
				binary.Read(rOutput, binary.LittleEndian, &vals)

				sampler.Value3 = vals
			}

			time := make([]float32, inputAccessor.Count)
			binary.Read(rInput, binary.LittleEndian, &time)

			sampler.Time = time
			sampler.Interpolation = docSampler.Interpolation
			sampler.Path = path

			animation.Samplers[i] = sampler
		}
		animations = append(animations, animation)
	}

	return animations
}
