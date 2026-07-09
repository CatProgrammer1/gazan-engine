package main

import (
	"fmt"
	"gl/yks"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/glfw/v3.4/glfw"
	"github.com/go-gl/mathgl/mgl32"

	_ "image/jpeg"
	_ "image/png"
)

func handle(err error) {
	if err != nil {
		panic(err)
	}
}

func checkGLError(where string) {
	if err := gl.GetError(); err != gl.NO_ERROR {
		panic(fmt.Sprintf("%s: OpenGL error 0x%X", where, err))
	}
}

var (
	mainGame *Game

	meshesToLoad = []string{
		"shit/hata.glb",
		"shit/cube_metal.glb",
		"shit/plate.glb",
		"shit/pig twerk fix 2.glb",
		"shit/kolt.glb",
		"shit/artem.glb",
	}

	workspace = make(map[string]Object)

	meshCache = make(map[uint32]*Mesh)

	CurrentTime,
	DeltaTime float32
)

const (
	scriptsDir = "scripts"

	vertexStride = int32(unsafe.Sizeof(Vertex{}))
)

func gamef() {
	hataObj, ok := workspace["hata"]
	if !ok {
		return
	}
	hata := hataObj.(*MeshObject)

	hata.SetPosition(mgl32.Vec3{-5, .1, 0})

	plateObj, ok := workspace["plate"]
	if !ok {
		return
	}
	plate := plateObj.(*MeshObject)

	plate.SetPosition(mgl32.Vec3{0, -1, 0})
	plate.SetScale(mgl32.Vec3{10, .1, 10})

	plateObj2, ok := workspace["cube_metal"]
	if !ok {
		return
	}
	plate2 := plateObj2.(*MeshObject)

	plate2.SetPosition(mgl32.Vec3{1, 0, 5})
	plate2.SetScale(mgl32.Vec3{.4, .4, .4})

	/*pigObj, ok := workspace["pig twerk fix 2"]
	if !ok {
		return
	}
	pig := pigObj.(*MeshObject)*/

	/*anim1 := pig.Animations[5]
	if !anim1.IsPlaying {
		anim1.Play(CurrentTime)
	}*/

	//pig.SetPosition(mgl32.Vec3{0, -.915, -5})

	artemObj, ok := workspace["artem"]
	if !ok {
		return
	}
	artem := artemObj.(*MeshObject)

	artem.SetPosition(mgl32.Vec3{0, -.915, -5})

	anim1 := artem.Animations[0]
	if !anim1.IsPlaying {
		anim1.Play(CurrentTime)
	}

	gearObj, ok := workspace["kolt"]
	if !ok {
		return
	}
	gear := gearObj.(*MeshObject)

	/*offsetX := float32(94.0580) / 100.0
	offsetY := float32(-6.4412) / 100.0
	offsetZ := float32(-99.1354) / 100.0 // Инвертируем Z, так как в glTF Forward — это -Z

	translationMatrix := mgl32.Translate3D(offsetX, offsetY, offsetZ)

	offsetMatrix := translationMatrix*/

	handWorldMatrix := artem.modelMatrix.Mul4(anim1.GetBone("mixamorig:RightHand").GlobalTransformation)

	gear.SetModelMatrix(resetScale(handWorldMatrix))

	/*anim2 := gear.Animations[0]

	if !anim2.IsPlaying {
		anim2.Play(CurrentTime)
	}*/

	// 2. Обертання (XYZ Euler)
	/*offsetX := float32(94.0580) / 100.0
	offsetY := float32(-6.4412) / 100.0
	offsetZ := float32(-99.1354) / 100.0 // Инвертируем Z, так как в glTF Forward — это -Z

	translationMatrix := mgl32.Translate3D(offsetX, offsetY, offsetZ)

	offsetMatrix := translationMatrix

	handWorldMatrix := pig.modelMatrix.Mul4(anim1.GetBone("mixamorig:RightHand").GlobalTransformation)

	// Применяем наш нулевой оффсет
	gear.SetModelMatrix(resetScale(handWorldMatrix.Mul4(offsetMatrix)))*/
}

/*
$env:CGO_CFLAGS="-I C:\msys64\mingw64\include"
$env:CGO_LDFLAGS="-L C:\msys64\mingw64\lib"
*/

func main() {

	runtime.LockOSThread()

	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	window, err := glfw.CreateWindow(1500, 1000, "Testing", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)

	err = gl.Init()
	if err != nil {
		panic(err)
	}

	skyboxTex := newCubeMapFromFile("shit/sky.png")

	postProcessShaderPrograms := []ShaderProgram{}

	///!!!!!!!!!!!!!!!!!!!!!!!!!!!
	///!!!!!!!!!!!!!!!!!!!!!!!!!!!
	///!!!!!!!!!!!!!!!!!!!!!!!!!!!
	///!!!!!!!!!!!!!!!!!!!!!!!!!!!
	///!!!!!!!!!!!!!!!!!!!!!!!!!!!
	///!!!!!!!!!!!!!!!!!!!!!!!!!!!
	///!!!!!!!!!!!!!!!!!!!!!!!!!!!
	///!!!!!!!!!!!!!!!!!!!!!!!!!!!

	vertexShaderDepth, ok := newShaderFromFile("shaders/depth.vs", gl.VERTEX_SHADER)
	if !ok {
		log.Fatalln("Vertex shader issue")
	}
	fragmentShaderDepth, ok := newShaderFromFile("shaders/depth.fs", gl.FRAGMENT_SHADER)
	if !ok {
		log.Fatalln("Fragment shader issue")
	}

	shaderProgramDepth, ok := newShaderProgram()
	if !ok {
		log.Fatalln("Shader program issue")
	}

	shaderProgramDepth.AttachShaders(true, vertexShaderDepth, fragmentShaderDepth)
	shaderProgramDepth.Link()

	///??????????????????????????
	///??????????????????????????
	///??????????????????????????
	///??????????????????????????
	///??????????????????????????
	///??????????????????????????

	vertexShaderDepthP, ok := newShaderFromFile("shaders/depth_p.vs", gl.VERTEX_SHADER)
	if !ok {
		log.Fatalln("Vertex shader issue")
	}

	geometryShaderDepthP, ok := newShaderFromFile("shaders/depth_p.gs", gl.GEOMETRY_SHADER)
	if !ok {
		log.Fatalln("Geometry shader issue")
	}

	fragmentShaderDepthP, ok := newShaderFromFile("shaders/depth_p.fs", gl.FRAGMENT_SHADER)
	if !ok {
		log.Fatalln("Fragment shader issue")
	}

	shaderProgramDepthP, ok := newShaderProgram()
	if !ok {
		log.Fatalln("Shader program issue")
	}

	shaderProgramDepthP.AttachShaders(true, vertexShaderDepthP, geometryShaderDepthP, fragmentShaderDepthP)
	shaderProgramDepthP.Link()

	///??????????????????????????
	///??????????????????????????
	///??????????????????????????
	///??????????????????????????
	///??????????????????????????
	///??????????????????????????

	ppVSShader, ok := newShaderFromFile("shaders/pp.vs", gl.VERTEX_SHADER)
	if !ok {
		log.Fatalln("Vertex shader issue")
	}

	ppFSShader, ok := newShaderFromFile("shaders/pp.fs", gl.FRAGMENT_SHADER)
	if !ok {
		log.Fatalln("Fragment shader issue")
	}

	ppProgram, ok := newShaderProgram()
	if !ok {
		log.Fatalln("Shader program issue")
	}

	ppProgram.AttachShaders(true, ppVSShader, ppFSShader)
	ppProgram.Link()

	postProcessShaderPrograms = append(postProcessShaderPrograms, ppProgram)

	///??????????????????????????
	///??????????????????????????
	///??????????????????????????
	///??????????????????????????
	///??????????????????????????
	///??????????????????????????

	vertexShader, ok := newShaderFromFile("shaders/main.vs", gl.VERTEX_SHADER)
	if !ok {
		log.Fatalln("Vertex shader issue")
	}

	fragmentShader, ok := newShaderFromFile("shaders/main.fs", gl.FRAGMENT_SHADER)
	if !ok {
		log.Fatalln("Fragment shader issue")
	}

	shaderProgram, ok := newShaderProgram()
	if !ok {
		log.Fatalln("Shader program issue")
	}

	shaderProgram.AttachShaders(false, vertexShader, fragmentShader)
	shaderProgram.Link()

	fragmentShader.Delete()

	shaderProgram.Use()

	///??????????????????????????
	///??????????????????????????
	///??????????????????????????
	///??????????????????????????
	///??????????????????????????
	///??????????????????????????

	vertexSkyShader, ok := newShaderFromFile("shaders/sky.vs", gl.VERTEX_SHADER)
	if !ok {
		log.Fatalln("Vertex shader issue")
	}

	fragmentSkyShader, ok := newShaderFromFile("shaders/sky.fs", gl.FRAGMENT_SHADER)
	if !ok {
		log.Fatalln("Fragment shader issue")
	}

	shaderSkyProgram, ok := newShaderProgram()
	if !ok {
		log.Fatalln("Shader program issue")
	}

	shaderSkyProgram.AttachShaders(true, vertexSkyShader, fragmentSkyShader)
	fragmentSkyShader.Delete()
	shaderSkyProgram.Link()

	shaderSkyProgram.Use()

	///!!!!!!!!!!!!!!!!!!!!!!!!!!!
	///!!!!!!!!!!!!!!!!!!!!!!!!!!!
	///!!!!!!!!!!!!!!!!!!!!!!!!!!!
	///!!!!!!!!!!!!!!!!!!!!!!!!!!!
	///!!!!!!!!!!!!!!!!!!!!!!!!!!!
	///!!!!!!!!!!!!!!!!!!!!!!!!!!!
	///!!!!!!!!!!!!!!!!!!!!!!!!!!!
	///!!!!!!!!!!!!!!!!!!!!!!!!!!!

	//skyViewLocation, skyProjectionLocation, cubeMapLocation := shaderSkyProgram.GetUniformLocation(ViewMatrixUniform), shaderSkyProgram.GetUniformLocation(ProjectionMatrixUniform), shaderSkyProgram.GetUniformLocation("skybox")

	skyVertices := []Vertex{
		// BACK (-Z)
		{[3]float32{-1.0, -1.0, -1.0}, [3]float32{0.0, 0.0, -1.0}, [2]float32{0, 0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{-1.0, 1.0, -1.0}, [3]float32{0.0, 0.0, -1.0}, [2]float32{0.0, 1.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{1.0, 1.0, -1.0}, [3]float32{0.0, 0.0, -1.0}, [2]float32{1.0, 1.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{1.0, 1.0, -1.0}, [3]float32{0.0, 0.0, -1.0}, [2]float32{1.0, 1.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{1.0, -1.0, -1.0}, [3]float32{0.0, 0.0, -1.0}, [2]float32{1.0, 0.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{-1.0, -1.0, -1.0}, [3]float32{0.0, 0.0, -1.0}, [2]float32{0.0, 0.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},

		// FRONT (+Z)
		{[3]float32{-1.0, -1.0, 1.0}, [3]float32{0.0, 0.0, 1.0}, [2]float32{0.0, 0.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{1.0, -1.0, 1.0}, [3]float32{0.0, 0.0, 1.0}, [2]float32{1.0, 0.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{1.0, 1.0, 1.0}, [3]float32{0.0, 0.0, 1.0}, [2]float32{1.0, 1.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{1.0, 1.0, 1.0}, [3]float32{0.0, 0.0, 1.0}, [2]float32{1.0, 1.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{-1.0, 1.0, 1.0}, [3]float32{0.0, 0.0, 1.0}, [2]float32{0.0, 1.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{-1.0, -1.0, 1.0}, [3]float32{0.0, 0.0, 1.0}, [2]float32{0.0, 0.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},

		// LEFT (-X)
		{[3]float32{-1.0, 1.0, 1.0}, [3]float32{-1.0, 0.0, 0.0}, [2]float32{1.0, 0.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{-1.0, 1.0, -1.0}, [3]float32{-1.0, 0.0, 0.0}, [2]float32{1.0, 1.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{-1.0, -1.0, -1.0}, [3]float32{-1.0, 0.0, 0.0}, [2]float32{0.0, 1.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{-1.0, -1.0, -1.0}, [3]float32{-1.0, 0.0, 0.0}, [2]float32{0.0, 1.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{-1.0, -1.0, 1.0}, [3]float32{-1.0, 0.0, 0.0}, [2]float32{0.0, 0.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{-1.0, 1.0, 1.0}, [3]float32{-1.0, 0.0, 0.0}, [2]float32{1.0, 0.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},

		// RIGHT (+X)
		{[3]float32{1.0, 1.0, -1.0}, [3]float32{1.0, 0.0, 0.0}, [2]float32{1.0, 0.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{1.0, 1.0, 1.0}, [3]float32{1.0, 0.0, 0.0}, [2]float32{0.0, 0.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{1.0, -1.0, 1.0}, [3]float32{1.0, 0.0, 0.0}, [2]float32{0.0, 1.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{1.0, -1.0, 1.0}, [3]float32{1.0, 0.0, 0.0}, [2]float32{0.0, 1.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{1.0, -1.0, -1.0}, [3]float32{1.0, 0.0, 0.0}, [2]float32{1.0, 1.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{1.0, 1.0, -1.0}, [3]float32{1.0, 0.0, 0.0}, [2]float32{1.0, 0.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},

		// BOTTOM (-Y)
		{[3]float32{-1.0, -1.0, -1.0}, [3]float32{0.0, -1.0, 0.0}, [2]float32{0.0, 1.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{1.0, -1.0, -1.0}, [3]float32{0.0, -1.0, 0.0}, [2]float32{1.0, 1.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{1.0, -1.0, 1.0}, [3]float32{0.0, -1.0, 0.0}, [2]float32{1.0, 0.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{1.0, -1.0, 1.0}, [3]float32{0.0, -1.0, 0.0}, [2]float32{1.0, 0.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{-1.0, -1.0, 1.0}, [3]float32{0.0, -1.0, 0.0}, [2]float32{0.0, 0.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{-1.0, -1.0, -1.0}, [3]float32{0.0, -1.0, 0.0}, [2]float32{0.0, 1.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},

		// TOP (+Y)
		{[3]float32{-1.0, 1.0, -1.0}, [3]float32{0.0, 1.0, 0.0}, [2]float32{0.0, 1.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{-1.0, 1.0, 1.0}, [3]float32{0.0, 1.0, 0.0}, [2]float32{0.0, 0.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{1.0, 1.0, 1.0}, [3]float32{0.0, 1.0, 0.0}, [2]float32{1.0, 0.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{1.0, 1.0, 1.0}, [3]float32{0.0, 1.0, 0.0}, [2]float32{1.0, 0.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{1.0, 1.0, -1.0}, [3]float32{0.0, 1.0, 0.0}, [2]float32{1.0, 1.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
		{[3]float32{-1.0, 1.0, -1.0}, [3]float32{0.0, 1.0, 0.0}, [2]float32{0.0, 1.0}, [4]uint8{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}, [4]float32{0, 0, 0, 0}},
	}

	initDefault()

	gl.FrontFace(gl.CCW)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LEQUAL)
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)

	var lastTime float32

	glfw.SwapInterval(0)

	skybox3D := newMesh(skyVertices, nil, nil, gl.STATIC_DRAW,
		Attribute{0, 3, gl.FLOAT, false, vertexStride, 0},
	)
	defer skybox3D.Delete()

	go func() {
		for _, meshPath := range meshesToLoad {
			loadedMesh := newMeshFromFile(meshPath, gl.STATIC_DRAW, true,
				Attribute{0, 3, gl.FLOAT, false, vertexStride, 0},
				Attribute{1, 3, gl.FLOAT, false, vertexStride, uintptr(3 * 4)},
				Attribute{2, 2, gl.FLOAT, false, vertexStride, uintptr(6 * 4)},
				Attribute{3, 4, gl.UNSIGNED_BYTE, false, vertexStride, uintptr(8 * 4)},
				Attribute{4, 4, gl.FLOAT, false, vertexStride, uintptr(8*4 + 4)},
				Attribute{5, 4, gl.FLOAT, false, vertexStride, uintptr(12*4 + 4)},
			)

			loadedMeshes <- loadedMesh
		}
	}()

	mainGame = newGame(window)

	//moveSpeed := float32(1)

	lightSources := makeLightContainer(
		/*newSpotLightSource(
			mgl32.Vec3{1, .1, 0},
			mgl32.Vec3{0, 0, -1},
			mgl32.Vec3{2, 2, 2},
			30,
			mgl32.DegToRad(20),
			mgl32.DegToRad(25),
		),*/
		newPointLightSource(mgl32.Vec3{0, 3, 0}, mgl32.Vec3{2, 2, 2}, 10),
		/*newSpotLightSource(
			mgl32.Vec3{0, .1, 0},
			mgl32.Vec3{-1, 0, 0},
			mgl32.Vec3{2, 2, 2},
			30,
			mgl32.DegToRad(20),
			mgl32.DegToRad(25),
		),*/
		//newDirectionalLightSource(mgl32.Vec3{-1, -1, .5}, mgl32.Vec3{2.0, 2.0, 2.0}),
	)

	//maxFPS := float32(1 / 120)

	maxShadowResolution := int32(1024 * 6)

	borderColor := [4]float32{1, 1, 1, 1}

	mainGame.ShadowMaps = []ShadowMap{
		newShadowMap(shaderProgramDepth, "shadowMapArray1", maxShadowResolution, 1, gl.TEXTURE10, borderColor),
		newShadowMapCubeMap(shaderProgramDepthP, "shadowMapArray3", maxShadowResolution/6, 1, gl.TEXTURE12),
		newShadowMap(shaderProgramDepth, "shadowMapArray2", maxShadowResolution/3, 10, gl.TEXTURE11, borderColor),
	}

	mainGame.ShaderPrograms = []ShaderProgram{
		shaderProgram,
	}
	mainGame.SpotLightSources = lightSources.GetType(2)
	mainGame.PointLightSources = lightSources.GetType(1)
	mainGame.DirLightSources = lightSources.GetType(0)

	mainGame.LightTypeShadowMapIndex[0] = 0
	mainGame.LightTypeShadowMapIndex[1] = 1
	mainGame.LightTypeShadowMapIndex[2] = 2

	mainWorkspace := &Workspace{
		Game: mainGame,

		UseShadowMaps: true,

		Enable: []uint32{
			gl.DEPTH_TEST,
			gl.CULL_FACE,
		},
		Disable: []uint32{
			gl.BLEND,
		},
		DepthMask:     true,
		DepthFunc:     gl.LESS,
		Factor:        0.0,
		Units:         0.0,
		CullFace:      gl.BACK,
		ShaderProgram: shaderProgram,
		Objects:       workspace,

		CustomUniforms: make(map[string]*CustomUniform),
	}

	mainWorkspace.SetCustomUniform("environment", skyboxTex)

	mainGame.Workspaces = append(mainGame.Workspaces, mainWorkspace)

	mainGame.Workspaces = append(mainGame.Workspaces, &Workspace{
		Game: mainGame,
		Enable: []uint32{
			gl.DEPTH_TEST,
		},
		Disable: []uint32{
			gl.CULL_FACE,
		},
		DepthMask:     false,
		DepthFunc:     gl.LEQUAL,
		CullFace:      gl.FRONT,
		Factor:        0.0,
		Units:         0.0,
		ShaderProgram: shaderSkyProgram,
		Objects: map[string]Object{
			"skybox": newCubeMapObject(skybox3D, skyboxTex),
		},

		CustomUniforms: make(map[string]*CustomUniform),
	})

	mainGame.PostProcess = make([]*PostProcess, len(postProcessShaderPrograms))

	scripts_entries, err := os.ReadDir(scriptsDir)
	handle(err)

	if len(scripts_entries) > 0 {
		lexer := yks.NewLexer("", "")
		parser := yks.NewParser("", nil)

		for _, entry := range scripts_entries {
			name := entry.Name()

			path := filepath.Join(scriptsDir, name)

			content, err := os.ReadFile(path)
			handle(err)

			lexer.CurrentFileName = name
			lexer.Source = string(content)

			lexer.LoadSourceChars()

			tokens := lexer.GetTokens()

			parser.Tokens = tokens
			parser.CurrentFileName = name

			ast := parser.AST()

			interpreter := yks.NewInterpreter(name, ast)
			interpreter.Complete(false, builtinVals)

			mainGame.Scripts = append(mainGame.Scripts, interpreter)
		}
	}

	w, h := window.GetSize()

	initPostProcessing(int32(w), int32(h))

	ppEnabled = false

	for i, shaderProgram := range postProcessShaderPrograms {
		if i == 0 {
			mainGame.PostProcess[i] = newPostProcess(shaderProgram, gl.TEXTURE0)
			continue
		}
		mainGame.PostProcess[i] = newPostProcess(shaderProgram, gl.TEXTURE0)
	}

	for !window.ShouldClose() {
		if w == 0 && h == 0 {
			glfw.WaitEvents()
		}
		CurrentTime = float32(glfw.GetTime())

		performOperations()

		select {
		case newLoadedMesh := <-loadedMeshes:
			fmt.Printf("::LOADED NEW MESH - '%s';\n", newLoadedMesh.Name)
			meshPath := newLoadedMesh.Name

			base := filepath.Base(meshPath)
			ext := filepath.Ext(meshPath)

			workspace[strings.TrimSuffix(base, ext)] = newMeshObject(newLoadedMesh, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{1, 1, 1}, mgl32.QuatIdent())

			meshCache[newLoadedMesh.VAO] = newLoadedMesh
		default:
		}

		DeltaTime = CurrentTime - lastTime
		lastTime = CurrentTime

		gamef()

		mainGame.Update()

		window.SwapBuffers()
		glfw.PollEvents()
	}
}
