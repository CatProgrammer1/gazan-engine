package main

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/qmuntal/gltf"
)

type BoneBind struct {
	HeadBone *BoneInfo
	Offset mgl32.Mat4
	ValidOffset bool
	Enabled  bool
}

func makeBindingToBone(head *BoneInfo) *BoneBind {
	return &BoneBind{
		ValidOffset: false,
		Offset: mgl32.Mat4{},
		HeadBone: head,
		Enabled: true,
	}
}

type BoneInfo struct {
	ID                  int
	OffsetMatrix        mgl32.Mat4
	GlobalTransformation mgl32.Mat4
	FinalTransformation mgl32.Mat4
}

func parseDocumentSkins(document *gltf.Document) (map[string]int, []*BoneInfo) {
	var bonesInfo []*BoneInfo
	boneInfoIndexMap := make(map[string]int)

	for _, skin := range document.Skins {
		bonesInfo = make([]*BoneInfo, len(skin.Joints))
		for jointIndex, jointNode := range skin.Joints {
			node := document.Nodes[jointNode]

			name := node.Name
			if name == "" {
				name = fmt.Sprintf("bone_%d", jointIndex)
			}

			offsetMatrix := getBindMatrix(document, skin, jointIndex)

			boneInfoIndexMap[name] = jointIndex

			bonesInfo[jointIndex] = &BoneInfo{
				ID:           jointNode,
				OffsetMatrix: offsetMatrix,
			}
		}
	}

	return boneInfoIndexMap, bonesInfo
}
