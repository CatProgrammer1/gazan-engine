package main

import (
	"gl/yks"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Object interface {
	Draw(shaderProgram ShaderProgram, camera *Camera)
	GetModelMatrix() mgl32.Mat4
	SyncWithScript()
}

func newMeshObject(mesh *Mesh, position, scale mgl32.Vec3, rotation mgl32.Quat) *MeshObject {

	meshObject := &MeshObject{
		Mesh: mesh,

		Position: position,
		Scale:    scale,
		Rotation: rotation,
		isDirty:  true,

		Animations: []*Animation{},

		modelMatrix: mgl32.Ident4(),
	}

	for _, anim := range mesh.Animations {
		meshObject.AddAnim(anim)
	}

	bonesInfo, boneIndexMap := mesh.CloneBoneInfo()

	meshObject.bonesInfo = bonesInfo
	meshObject.boneIndexMap = boneIndexMap

	return meshObject
}

type MeshObject struct {
	Mesh *Mesh

	Animations []*Animation

	Position, Scale mgl32.Vec3
	Rotation        mgl32.Quat
	isDirty         bool

	boneIndexMap map[string]int
	bonesInfo    []*BoneInfo

	modelMatrix mgl32.Mat4

	ScriptMeshObject *yks.StructObject
}

// Clones the animation creating an unique copy and then adds it to the MeshObject.Animations slice
func (mesh *MeshObject) AddAnim(anim *Animation) {
	parentedAnim := new(Animation)

	parentedAnim.Channels = anim.Channels
	parentedAnim.Samplers = anim.Samplers
	parentedAnim.Document = anim.Document
	parentedAnim.IsPlaying = false
	parentedAnim.Looped = anim.Looped
	parentedAnim.LastTime = 0
	parentedAnim.Mesh = anim.Mesh
	parentedAnim.MeshObject = mesh
	parentedAnim.Name = anim.Name
	parentedAnim.Transforms = []mgl32.Mat4{}

	mesh.Animations = append(mesh.Animations, parentedAnim)
}

func (mesh *MeshObject) SyncWithScript() {
	scriptMeshObject := mesh.ScriptMeshObject
	if scriptMeshObject == nil || !scriptMeshObject.IsDirty {
		return
	}
	scriptMeshObject.IsDirty = false

	fx, fy, fz := [2]string{"X", "f32"},
		[2]string{"Y", "f32"},
		[2]string{"Z", "f32"}

	position := sigmaMustAssert[*yks.StructObject](scriptMeshObject.Get("Position"))
	position.CheckFormat(
		fx,
		fy,
		fz,
	)

	posX, posY, posZ := sigmaMustAssert[float32](position.Get("X")),
		sigmaMustAssert[float32](position.Get("Y")),
		sigmaMustAssert[float32](position.Get("Z"))

	scale := sigmaMustAssert[*yks.StructObject](scriptMeshObject.Get("Scale"))
	scale.CheckFormat(
		fx,
		fy,
		fz,
	)

	scaleX, scaleY, scaleZ := sigmaMustAssert[float32](scale.Get("X")),
		sigmaMustAssert[float32](scale.Get("Y")),
		sigmaMustAssert[float32](scale.Get("Z"))

	rotation := sigmaMustAssert[*yks.StructObject](scriptMeshObject.Get("Rotation"))
	rotation.CheckFormat(
		[2]string{"V", "Vec3"},
		[2]string{"W", "f32"},
	)

	rotW := sigmaMustAssert[float32](rotation.Get("W"))
	rotVec3 := sigmaMustAssert[*yks.StructObject](rotation.Get("V"))
	rotVec3.CheckFormat(
		fx,
		fy,
		fz,
	)

	rotX, rotY, rotZ := sigmaMustAssert[float32](rotVec3.Get("X")),
		sigmaMustAssert[float32](rotVec3.Get("Y")),
		sigmaMustAssert[float32](rotVec3.Get("Z"))

	mesh.Position = mgl32.Vec3{posX, posY, posZ}
	mesh.Scale = mgl32.Vec3{scaleX, scaleY, scaleZ}
	mesh.Rotation = mgl32.Quat{
		V: mgl32.Vec3{rotX, rotY, rotZ},
		W: rotW,
	}
}

var staticIdentities [100]mgl32.Mat4

func (mesh *MeshObject) Draw(shaderProgram ShaderProgram, camera *Camera) {
	modelLocation, viewLocation, projectionLocation :=
		shaderProgram.GetUniformLocation(ModelMatrixUniform),
		shaderProgram.GetUniformLocation(ViewMatrixUniform),
		shaderProgram.GetUniformLocation(ProjectionMatrixUniform)
	gBonesLocation := shaderProgram.GetUniformLocation(GBonesUniform)

	model := mesh.GetModelMatrix()

	for _, anim := range mesh.Animations {
		anim.SyncWithScript()

		if anim.IsPlaying {
			anim.Update(CurrentTime)
		}
	}

	var transforms []mgl32.Mat4
	for _, anim := range mesh.Animations {
		if anim.IsPlaying {
			transforms = anim.Transforms
		}
	}

	if len(transforms) > 0 {
		gl.UniformMatrix4fv(gBonesLocation, int32(len(transforms)), false, &transforms[0][0])
	} else {
		gl.UniformMatrix4fv(gBonesLocation, 100, false, &staticIdentities[0][0])
	}

	gl.UniformMatrix4fv(modelLocation, 1, false, &model[0])
	gl.UniformMatrix4fv(viewLocation, 1, false, &camera.view[0])
	gl.UniformMatrix4fv(projectionLocation, 1, false, &camera.projection[0])

	mesh.Mesh.DrawElements(shaderProgram, gl.TRIANGLES, gl.UNSIGNED_INT)
}

func (mesh *MeshObject) DrawShadow(shaderProgramDepth ShaderProgram) {
	modelLocation := shaderProgramDepth.GetUniformLocation(ModelMatrixUniform)
	gBonesLocation := shaderProgramDepth.GetUniformLocation(GBonesUniform)

	model := mesh.GetModelMatrix()

	for _, anim := range mesh.Animations {
		anim.SyncWithScript()

		if anim.IsPlaying {
			anim.Update(CurrentTime)
		}
	}
	var transforms []mgl32.Mat4
	for _, anim := range mesh.Animations {
		if anim.IsPlaying {
			transforms = anim.Transforms
		}
	}
	if len(transforms) > 0 {
		gl.UniformMatrix4fv(gBonesLocation, int32(len(transforms)), false, &transforms[0][0])
	} else {
		gl.UniformMatrix4fv(gBonesLocation, 100, false, &staticIdentities[0][0])
	}

	gl.UniformMatrix4fv(modelLocation, 1, false, &model[0])

	mesh.Mesh.DrawElements(shaderProgramDepth, gl.TRIANGLES, gl.UNSIGNED_INT)
}

func (mesh *MeshObject) GetModelMatrix() mgl32.Mat4 {
	if mesh.isDirty {
		trans := mgl32.Translate3D(mesh.Position[0], mesh.Position[1], mesh.Position[2])
		rotate := mesh.Rotation.Mat4()
		scale := mgl32.Scale3D(mesh.Scale[0], mesh.Scale[1], mesh.Scale[2])

		mesh.isDirty = false
		mesh.modelMatrix = trans.Mul4(rotate).Mul4(scale)
	}
	return mesh.modelMatrix
}

// Completely overrides the previous model matrix, change this if you at least know the basics of matrix math
func (mesh *MeshObject) SetModelMatrix(mat4 mgl32.Mat4) {
	mesh.modelMatrix = mat4
}

func (mesh *MeshObject) SetPosition(v3 mgl32.Vec3) {
	isDirty := !mesh.Position.ApproxEqual(v3)
	if isDirty {
		mesh.isDirty = isDirty
		mesh.Position = v3
	}
}

func (mesh *MeshObject) SetScale(v3 mgl32.Vec3) {
	isDirty := !mesh.Scale.ApproxEqual(v3)
	if isDirty {
		mesh.isDirty = isDirty
		mesh.Scale = v3
	}
}

func (mesh *MeshObject) SetRotation(q mgl32.Quat) {
	isDirty := !mesh.Rotation.ApproxEqual(q)
	if isDirty {
		mesh.isDirty = isDirty
		mesh.Rotation = q
	}
}

func (mesh *MeshObject) GPUInstanceData() GPUInstanceData {
	gbones := [100]mgl32.Mat4{}

	for i := 0; i < 100; i++ {
		gbones[i] = mesh.bonesInfo[i].FinalTransformation
	}

	return GPUInstanceData{
		ModelMatrix: mesh.GetModelMatrix(),
		GBones:      gbones,
	}
}

type GPUInstanceData struct {
	ModelMatrix mgl32.Mat4
	GBones      [100]mgl32.Mat4
}

func newCubeMapObject(mesh *Mesh, cubeMap *CubeMap) *CubeMapObject {
	return &CubeMapObject{
		Mesh:    mesh,
		CubeMap: cubeMap,
	}
}

type CubeMapObject struct {
	Mesh    *Mesh
	CubeMap *CubeMap
}

func (cubeMapObj *CubeMapObject) Draw(shaderProgram ShaderProgram, camera *Camera) {
	cubeMap := cubeMapObj.CubeMap
	mesh := cubeMapObj.Mesh

	viewLocation, projectionLocation :=
		shaderProgram.GetUniformLocation(ViewMatrixUniform),
		shaderProgram.GetUniformLocation(ProjectionMatrixUniform)

	view := camera.view.Mat3().Mat4()

	gl.UniformMatrix4fv(viewLocation, 1, false, &view[0])
	gl.UniformMatrix4fv(projectionLocation, 1, false, &camera.projection[0])

	cubeMap.Bind(shaderProgram.GetUniformLocation(SkyboxUniform))

	mesh.DrawArrays(gl.TRIANGLES, 0, int32(len(mesh.Vertices)))
}

func (cubeMapObj *CubeMapObject) SyncWithScript() {

}

func (cubeMapObj *CubeMapObject) GetModelMatrix() mgl32.Mat4 {
	return mgl32.Ident4()
}

/*
type MeshObjectsGroup struct {
	Mesh *Mesh

	SubmeshObjectsGroup []*SubmeshObjectsGroup
}

func newMeshObjectsGroup(mesh *Mesh) *MeshObjectsGroup {
	return &MeshObjectsGroup{
		Mesh: mesh,
		SubmeshObjectsGroup: []*SubmeshObjectsGroup{},
	}
}

func (mog *MeshObjectsGroup) AddObject(mobj *MeshObject) {
	mog.Objects = append(mog.Objects, mobj)
}

type SubmeshObjectsGroup struct {
	MeshObjectsGroup *MeshObjectsGroup

	SSBO uint32
	Objects []*MeshObject
}

func (mog *MeshObjectsGroup) Draw(shaderProgram ShaderProgram, camera *Camera) {
	if len(mog.Objects) == 0 {
		return
	}

	mesh := mog.Mesh

	mog.GPUInstancesData = mog.GPUInstancesData[:0]

	for _, obj := range mog.Objects {
		mog.GPUInstancesData = append(mog.GPUInstancesData, obj.GPUInstanceData())
	}

	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, mog.SSBO)
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 0, mog.SSBO)

	size := len(mog.GPUInstancesData) * int(unsafe.Sizeof(GPUInstanceData{}))

	gl.BufferData(gl.SHADER_STORAGE_BUFFER, size, unsafe.Pointer(&mog.GPUInstancesData[0]), gl.STREAM_DRAW)

	shaderProgram.Use()

	gl.BindVertexArray(mesh.VAO)


	for _, submesh := range mesh.SubMeshes {
		if submesh.Material != nil {
			submesh.Material.Use(shaderProgram, "material")
		} else if defaultMaterial != nil {
			defaultMaterial.Use(shaderProgram, "material")
		}
		gl.DrawElementsInstanced(gl.TRIANGLES, submesh, gl.UNSIGNED_INT, unsafe.Pointer(&mesh.Indices[0]), int32(len(mog.Objects)))
	}
	gl.DrawElementsInstanced(gl.TRIANGLES, int32(len(mesh.Indices)), gl.UNSIGNED_INT, unsafe.Pointer(&mesh.Indices[0]), int32(len(mog.Objects)))
}
*/
