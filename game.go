package main

import (
	"fmt"
	"gl/yks"
	"log"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/glfw/v3.4/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type CustomUniform struct {
	Location int32
	Value    any
	isDirty  bool
}

func (cu *CustomUniform) Set(v any) {
	cu.isDirty = true
	cu.Value = v
}

type Workspace struct {
	Name string

	UseShadowMaps bool
	Game          *Game
	//Enables all the given flags at the beginning when Workspace is rendered, and disables them at the end
	Enable []uint32

	//Disables all the given flags at the beginning when Workspace is rendered, and enables them at the end
	Disable []uint32

	DepthMask bool
	DepthFunc uint32

	//PolygonOffset
	Factor, Units float32

	CullFace uint32

	CustomUniforms map[string]*CustomUniform

	ShaderProgram ShaderProgram
	Objects       map[string]Object

	ScriptWorkspace *yks.StructObject
}

func (workspace *Workspace) SetCustomUniform(name string, value any) {
	if customUniform, ok := workspace.CustomUniforms[name]; ok {
		customUniform.isDirty = true
		customUniform.Value = value
		return
	}

	workspace.CustomUniforms[name] = &CustomUniform{
		isDirty:  true,
		Location: workspace.ShaderProgram.GetUniformLocation(name),
		Value:    value,
	}
}

func (workspace *Workspace) SetUniform(uniform int32, value any) {
	workspace.ShaderProgram.SetUniform(uniform, value)
}

func (workspace *Workspace) AttachShaderProgram(shaderProgram ShaderProgram) {
	workspace.ShaderProgram = shaderProgram
}

func (workspace *Workspace) DrawFBO(shadowMap ShadowMap) {
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthMask(true)
	gl.Enable(gl.POLYGON_OFFSET_FILL)
	gl.PolygonOffset(.15, 2)
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)

	for _, object := range workspace.Objects {
		if meshObj, ok := object.(*MeshObject); ok {
			meshObj.DrawShadow(shadowMap.ShaderProgram)
		}
	}

	gl.CullFace(gl.FRONT)
	gl.Disable(gl.POLYGON_OFFSET_FILL)
	gl.Disable(gl.CULL_FACE)
}

func (workspace *Workspace) DrawObjects(camera *Camera) {
	shaderProgram := workspace.ShaderProgram

	shaderProgram.Use()

	////////////////////////////////////

	if workspace.UseShadowMaps {
		for _, shadowMap := range workspace.Game.ShadowMaps {
			gl.ActiveTexture(shadowMap.TextureUnit)
			gl.BindTexture(shadowMap.Target, shadowMap.DepthMap)

			shadowMap.SetUniform(shaderProgram)
		}
	}

	i := 0
	for _, lightSource := range workspace.Game.SpotLightSources {
		lightSource.UpdateAttenuationCoefficients()
		lightSource.SetUniform(shaderProgram, fmt.Sprintf("lightSources[%d]", i))
		i++
	}

	for _, lightSource := range workspace.Game.PointLightSources {
		lightSource.UpdateAttenuationCoefficients()
		lightSource.SetUniform(shaderProgram, fmt.Sprintf("lightSources[%d]", i))
		i++
	}

	for _, lightSource := range workspace.Game.DirLightSources {
		lightSource.UpdateAttenuationCoefficients()
		lightSource.SetUniform(shaderProgram, fmt.Sprintf("lightSources[%d]", i))
		i++
	}

	for _, customUniform := range workspace.CustomUniforms {
		if !customUniform.isDirty {
			continue
		}
		customUniform.isDirty = false
		shaderProgram.SetUniform(customUniform.Location, customUniform.Value)
	}

	gl.Uniform1i(shaderProgram.GetUniformLocation(LightSourcesCountUniform), int32(len(workspace.Game.SpotLightSources)+len(workspace.Game.DirLightSources)+len(workspace.Game.PointLightSources)))

	gl.Uniform3f(shaderProgram.GetUniformLocation(ViewPositionUniform), camera.Position[0], camera.Position[1], camera.Position[2])
	if workspace.CullFace == 0 {
		workspace.CullFace = gl.BACK
	}
	gl.CullFace(workspace.CullFace)

	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.PolygonOffset(workspace.Factor, workspace.Units)

	for _, e_flag := range workspace.Enable {
		gl.Enable(e_flag)
	}

	for _, d_flag := range workspace.Disable {
		gl.Disable(d_flag)
	}

	gl.DepthFunc(workspace.DepthFunc)

	gl.DepthMask(workspace.DepthMask)

	for _, object := range workspace.Objects {
		object.SyncWithScript()

		object.Draw(shaderProgram, camera)
	}

	for _, e_flag := range workspace.Enable {
		gl.Disable(e_flag)
	}
}

func (workspace *Workspace) GetObject(name string) Object {
	return workspace.Objects[name]
}

func newGame(window *glfw.Window) *Game {
	game := &Game{
		window: window,

		LightTypeShadowMapIndex: make(map[int32]int),
		DirLightSources:         []*Light{},
		SpotLightSources:        []*Light{},
		PointLightSources:       []*Light{},

		ShadowMaps:     []ShadowMap{},
		ShaderPrograms: []ShaderProgram{},
		Workspaces:     []*Workspace{},

		ListenKeys:    []int64{},
		ListenButtons: []int64{},

		Meshes:  []*Mesh{},
		Objects: make(map[string]Object),

		PostProcess: []*PostProcess{},

		sm_indeces: make(map[int32]int32),
	}

	w, h := window.GetSize()

	camera := newCamera(window, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 0, -1}, mgl32.Vec3{0, 1, 0}, mgl32.Perspective(mgl32.DegToRad(45), float32(w)/float32(h), .01, 100))

	game.Camera = camera

	return game
}

type Game struct {
	window *glfw.Window

	Camera *Camera

	ListenKeys    []int64
	ListenButtons []int64

	ShadowMaps     []ShadowMap
	ShaderPrograms []ShaderProgram

	PostProcess []*PostProcess

	LightTypeShadowMapIndex map[int32]int

	//LightSources []*Light
	SpotLightSources  []*Light
	DirLightSources   []*Light
	PointLightSources []*Light

	//To make the object appear in the world you must add it to the workspace, Meshes slice and Objects map are just used to store the instances to get it if needed withouth searching through workspaces
	Meshes  []*Mesh
	Objects map[string]Object

	Workspaces []*Workspace

	Scripts []*yks.Interpreter

	sm_indeces map[int32]int32
}

func (game *Game) AddLightSrc(light *Light) {
	switch light.Type {
	case 0:
		game.DirLightSources = append(game.DirLightSources, light)
	case 1:
		game.PointLightSources = append(game.PointLightSources, light)
	case 2:
		game.SpotLightSources = append(game.SpotLightSources, light)
	default:
		warn("Unexisting light source type. Light source wasn't added to the list.")
	}
}

func (game *Game) AddWorkspace(workspace *Workspace) {
	game.Workspaces = append(game.Workspaces, workspace)
}

func (game *Game) GetWorkspace(name string) *Workspace {
	for _, workspace := range game.Workspaces {
		if workspace.Name == name {
			return workspace
		}
	}
	return nil
}

func (game *Game) AddShadowMap(shadowMap ShadowMap) {
	game.ShadowMaps = append(game.ShadowMaps, shadowMap)
}

func (game *Game) AddShaderProgram(shaderProgram ShaderProgram) {
	game.ShaderPrograms = append(game.ShaderPrograms, shaderProgram)
}

func (game *Game) AddMesh(mesh *Mesh) {
	game.Meshes = append(game.Meshes, mesh)
}

func (game *Game) AddObject(name string, obj Object) {
	if _, ok := game.Objects[name]; ok {
		warn(fmt.Sprintf("Object with name '%s' already exists. Object was not added to a game storage. Expect issues with it not showing up correctly.", name))
	}
	game.Objects[name] = obj
}

func (game *Game) GetMesh(name string) *Mesh {
	for _, mesh := range game.Meshes {
		if mesh.Name == name {
			return mesh
		}
	}
	return nil
}

var funcCallTemp *yks.FuncCall = &yks.FuncCall{
	Arguments: make([]yks.Node, 1),
}

var (
	argumentsTemplates = make([][]yks.Node, 10)
)

func (game *Game) Update() {
	window := game.window

	for _, script := range game.Scripts {
		update, ok := script.CurrentScope.Data["Update"]
		if ok && update.FuncValue != nil {
			updateFuncDec := update.FuncValue

			funcCallTemp.Arguments = argumentsTemplates[1]

			funcCallTemp.Arguments[0] = &yks.FloatNode{
				Value: float64(DeltaTime),
				X:     0, Y: 0,
			}
			funcCallTemp.Func = updateFuncDec

			script.CompleteNode(funcCallTemp)
		}

		onInputEvent, ok := script.CurrentScope.Data["OnInputEvent"]
		if ok && onInputEvent.FuncValue != nil {
			onInputEventFuncDec := onInputEvent.FuncValue

			funcCallTemp.Arguments = argumentsTemplates[3]
			funcCallTemp.Func = onInputEventFuncDec

			for _, key := range game.ListenKeys {
				action := window.GetKey(glfw.Key(key))

				funcCallTemp.Arguments[0] = &yks.IntNode{
					ValueI64: yks.Rawint64(key),
					X:        0, Y: 0,
				}
				funcCallTemp.Arguments[1] = &yks.IntNode{
					ValueI64: yks.Rawint64(action),
					X:        0, Y: 0,
				}
				funcCallTemp.Arguments[2] = &yks.BoolNode{
					Value: false,
					X:     0, Y: 0,
				}

				script.CompleteNode(funcCallTemp)
			}

			for _, button := range game.ListenButtons {
				action := window.GetMouseButton(glfw.MouseButton(button))

				funcCallTemp.Arguments[0] = &yks.IntNode{
					ValueI64: yks.Rawint64(button),
					X:        0, Y: 0,
				}
				funcCallTemp.Arguments[1] = &yks.IntNode{
					ValueI64: yks.Rawint64(action),
					X:        0, Y: 0,
				}
				funcCallTemp.Arguments[2] = &yks.BoolNode{
					Value: true,
					X:     0, Y: 0,
				}

				script.CompleteNode(funcCallTemp)
			}
		}
	}

	camera := game.Camera

	camera.Update()

	camera.SyncWithScript()

	clear(game.sm_indeces)

	w, h := game.window.GetSize()

	dirLightShadowMapINDEX, ok := game.LightTypeShadowMapIndex[0]
	if ok && dirLightShadowMapINDEX < len(game.ShadowMaps) {
		dirLightShadowMap := game.ShadowMaps[dirLightShadowMapINDEX]

		dirLightShadowMap.Bind()
		gl.Clear(gl.DEPTH_BUFFER_BIT)

		for i, lightSource := range game.DirLightSources {
			if i+1 > int(dirLightShadowMap.Layers) {
				log.Printf("Too many light sources - %d, and too few layers in shadow map - %d.\n", len(game.DirLightSources), dirLightShadowMap.Layers)
				break
			}
			lightSource.SyncWithScript()
			lightSource.UpdateLightSpaceMatrix(camera)

			gl.UniformMatrix4fv(dirLightShadowMap.ShaderProgram.GetUniformLocation(LightSpaceMatrix), 1, false, &lightSource.LightSpaceMatrix[0][0])

			gl.FramebufferTextureLayer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, dirLightShadowMap.DepthMap, 0, int32(i))

			gl.Clear(gl.DEPTH_BUFFER_BIT)

			for _, workspace := range game.Workspaces {
				if !workspace.UseShadowMaps {
					continue
				}
				workspace.DrawFBO(dirLightShadowMap)
			}

		}
	}

	pointLightShadowMapINDEX, ok := game.LightTypeShadowMapIndex[1]
	if ok && pointLightShadowMapINDEX < len(game.ShadowMaps) {
		pointLightShadowMap := game.ShadowMaps[pointLightShadowMapINDEX]

		pointLightShadowMap.Bind()
		gl.Clear(gl.DEPTH_BUFFER_BIT)

		smShaderProgram := pointLightShadowMap.ShaderProgram

		gl.FramebufferTexture(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, pointLightShadowMap.DepthMap, 0)

		for i, lightSource := range game.PointLightSources {
			if i+1 > int(pointLightShadowMap.Layers) {
				log.Printf("Too many light sources - %d, and too few layers in shadow map - %d.\n", len(game.PointLightSources), pointLightShadowMap.Layers)
				break
			}
			lightSource.SyncWithScript()
			lightSource.UpdateLightSpaceMatrix(camera)

			gl.Uniform1f(smShaderProgram.GetUniformLocation("far_plane"), lightSource.MaxDistance)
			gl.Uniform3f(smShaderProgram.GetUniformLocation("lightPos"), lightSource.Position[0], lightSource.Position[1], lightSource.Position[2])
			gl.Uniform1i(smShaderProgram.GetUniformLocation("cubeIndex"), int32(i))
			gl.UniformMatrix4fv(smShaderProgram.GetUniformLocation(LightSpaceMatrix), 6, false, &lightSource.LightSpaceMatrix[0][0])

			for _, workspace := range game.Workspaces {
				if !workspace.UseShadowMaps {
					continue
				}
				workspace.DrawFBO(pointLightShadowMap)
			}

		}
	}

	spotLightShadowMapINDEX, ok := game.LightTypeShadowMapIndex[2]
	if ok && spotLightShadowMapINDEX < len(game.ShadowMaps) {
		spotLightShadowMap := game.ShadowMaps[spotLightShadowMapINDEX]

		spotLightShadowMap.Bind()
		gl.Clear(gl.DEPTH_BUFFER_BIT)

		for i, lightSource := range game.SpotLightSources {
			if i+1 > int(spotLightShadowMap.Layers) {
				log.Printf("Too many light sources - %d, and too few layers in shadow map - %d.\n", len(game.SpotLightSources), spotLightShadowMap.Layers)
				break
			}
			lightSource.SyncWithScript()
			lightSource.UpdateLightSpaceMatrix(camera)

			gl.UniformMatrix4fv(spotLightShadowMap.ShaderProgram.GetUniformLocation(LightSpaceMatrix), 1, false, &lightSource.LightSpaceMatrix[0][0])

			gl.FramebufferTextureLayer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, spotLightShadowMap.DepthMap, 0, int32(i))

			gl.Clear(gl.DEPTH_BUFFER_BIT)

			for _, workspace := range game.Workspaces {
				if !workspace.UseShadowMaps {
					continue
				}
				workspace.DrawFBO(spotLightShadowMap)
			}

		}
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Viewport(0, 0, int32(w), int32(h))

	gl.ClearColor(0.01, 0.01, 0.01, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	if ppEnabled {
		gl.BindFramebuffer(gl.FRAMEBUFFER, mainPPFBO)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	}

	for _, workspace := range game.Workspaces {
		workspace.DrawObjects(game.Camera)
	}

	if ppEnabled {
		gl.Disable(gl.DEPTH_TEST)
		gl.Disable(gl.CULL_FACE)

		currentSourceTexture := mainPPTexture
		lastIndex := len(game.PostProcess) - 1

		for i, postProcess := range game.PostProcess {
			postProcess.ShaderProgram.Use()

			if i == lastIndex {
				gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
			} else {
				gl.BindFramebuffer(gl.FRAMEBUFFER, ppFBOs[i%2])
			}

			gl.Clear(gl.COLOR_BUFFER_BIT)

			gl.ActiveTexture(postProcess.TextureUnit)
			gl.BindTexture(gl.TEXTURE_2D, currentSourceTexture)

			postProcess.Bind("frame_image")

			quadMesh.DrawArrays(gl.TRIANGLES, 0, int32(len(quadMesh.Vertices)))

			if i != lastIndex {
				currentSourceTexture = ppTextures[i%2]
			}
		}
	}
}
