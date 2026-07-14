package main

import (
	"errors"
	"fmt"
	"gl/yks"
	"log"
	"math"
	"runtime"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

func makeStructObjectFromStructure(structure *yks.Structure, fields map[string]*yks.Field) *yks.StructObject {
	structObject := &yks.StructObject{
		Identifier: structure.Identifier,

		Fields:  fields,
		Methods: make(map[string]*yks.Method, structure.CountMethods()),

		LastMem: []byte{},
	}

	for _, field := range structure.Fields {
		if field.Method {

			fieldDeclFunc := field.Func

			methodFuncClone := new(yks.FuncDec)
			methodFuncClone.Self = structObject
			methodFuncClone.Arguments = fieldDeclFunc.Arguments
			methodFuncClone.ArgumentsDataTypes = fieldDeclFunc.ArgumentsDataTypes
			methodFuncClone.Body = fieldDeclFunc.Body
			methodFuncClone.Identifier = fieldDeclFunc.Identifier
			methodFuncClone.ReturnDataTypes = fieldDeclFunc.ReturnDataTypes
			methodFuncClone.Template = fieldDeclFunc.Template
			methodFuncClone.X = fieldDeclFunc.X
			methodFuncClone.Y = fieldDeclFunc.Y

			structObject.Methods[field.Identifier] = &yks.Method{
				Identifier: field.Identifier,
				Func: &yks.Cell{
					DataType:  "func",
					FuncValue: methodFuncClone,
				},
			}
		}
	}

	return structObject
}

func sigmaAssertB[T any](v any, ok bool) (T, bool) {
	t, suc := v.(T)
	if !suc {
		return t, false
	}

	return t, ok
}

func sigmaAssert[T any](v any, ok bool) (T, bool) {
	t, suc := v.(T)
	if !suc {
		panic("nga damn")
	}

	return t, ok
}

func sigmaMustAssert[T any](v any, ok bool) T {
	t, suc := v.(T)
	if !suc {
		panic("nga damn")
	}

	if !ok {
		panic("bro shysh")
	}

	return t
}

var (
	gameYKSStructure = &yks.Structure{
		Identifier: "Game",
		Fields: []*yks.FieldDecl{
			{
				Identifier: "AddLightSrc",
				DataType:   "func",
				Method:     true,

				//* AddLightSrc
				Func: yks.NewFTemp("AddLightSrc", func(v ...any) []any {
					yks.ArgsCheck(v, 1, 1, "Light")

					x, y := v[0].(int), v[1].(int)
					inter := v[2].(*yks.Interpreter)

					v = v[yks.BUILTIN_SPECIALS:]

					light := v[0].(*yks.StructObject)

					ok, reason := light.CheckFormat(
						[2]string{"Type", "i32"},

						[2]string{"Position", "Vec3"},
						[2]string{"Direction", "Vec3"},
						[2]string{"Diffuse", "Vec3"},

						[2]string{"MaxDistance", "f32"},

						[2]string{"InnerCutOut", "f32"},
						[2]string{"OuterCutOut", "f32"},
					)
					if !ok {
						yks.Throw(inter.CurrentFileName, reason, x, y)
					}

					ltype := sigmaMustAssert[int32](light.Get("Type"))

					light.IsDirty = true

					lightSrc := newLightSource(ltype,
						//?Diffuse:
						mgl32.Vec3{},
						//?Position:
						mgl32.Vec3{},
						//?Direction:
						mgl32.Vec3{},
						//?MaxDistance:
						0,
						//?CutOuts:
						0, 0,
					)
					lightSrc.ScriptLight = light

					lightSrc.SyncWithScript()

					mainGame.AddLightSrc(lightSrc)

					return []any{}
				}),
			},
			{
				Identifier: "AddMesh",
				DataType:   "func",
				Method:     true,

				//* AddMesh
				Func: yks.NewFTemp("AddMesh", func(v ...any) []any {
					yks.ArgsCheck(v, 1, 1, "Mesh")

					x, y := v[0].(int), v[1].(int)
					inter := v[2].(*yks.Interpreter)

					v = v[yks.BUILTIN_SPECIALS:]

					mesh := v[0].(*yks.StructObject)

					ok, reason := mesh.CheckFormat(
						[2]string{"Name", "string"},
					)
					if !ok {
						yks.Throw(inter.CurrentFileName, reason, x, y)
					}

					name := sigmaMustAssert[string](mesh.Get("Name"))

					mesh.IsDirty = true

					origMesh := mainGame.GetMesh(name)

					mainGame.AddMesh(origMesh)

					return []any{}
				}),
			},
			{
				Identifier: "AddWorkspace",
				DataType:   "func",
				Method:     true,

				//* AddWorkspace
				Func: yks.NewFTemp("AddWorkspace", func(v ...any) []any {
					yks.ArgsCheck(v, 1, 1, "Workspace")

					x, y := v[0].(int), v[1].(int)
					inter := v[2].(*yks.Interpreter)

					v = v[yks.BUILTIN_SPECIALS:]

					workspaceObj := v[0].(*yks.StructObject)

					ok, reason := workspaceObj.CheckFormat(
						[2]string{"Name", "string"},

						[2]string{"UseShadowMaps", "bool"},

						[2]string{"Enable", "table"},
						[2]string{"Disable", "table"},

						[2]string{"DepthMask", "bool"},
						[2]string{"DepthFunc", "u32"},

						[2]string{"Factor", "f32"},
						[2]string{"Units", "f32"},

						[2]string{"CullFace", "u32"},

						[2]string{"ShaderProgram", "ShaderProgram"},

						[2]string{"Objects", "table"},
					)
					if !ok {
						yks.Throw(inter.CurrentFileName, reason, x, y)
					}

					name := sigmaMustAssert[string](workspaceObj.Get("Name"))

					useShadowMaps := sigmaMustAssert[bool](workspaceObj.Get("UseShadowMaps"))

					depthMask := sigmaMustAssert[bool](workspaceObj.Get("DepthMask"))
					depthFunc := sigmaMustAssert[uint32](workspaceObj.Get("DepthFunc"))

					factor := sigmaMustAssert[float32](workspaceObj.Get("Factor"))
					units := sigmaMustAssert[float32](workspaceObj.Get("Units"))

					cullFace := sigmaMustAssert[uint32](workspaceObj.Get("CullFace"))

					enableMap := sigmaMustAssert[*yks.Map](workspaceObj.Get("Enable"))
					if enableMap.DataType != "u32" {
						yks.Throw(inter.CurrentFileName, "Field Enable must be a table with datatype 'u32', not '%s'", x, y, enableMap.DataType)
					}

					disableMap := sigmaMustAssert[*yks.Map](workspaceObj.Get("Disable"))
					if disableMap.DataType != "u32" {
						yks.Throw(inter.CurrentFileName, "Field Disable must be a table with datatype 'u32', not '%s'", x, y, disableMap.DataType)
					}

					objectsMap := sigmaMustAssert[*yks.Map](workspaceObj.Get("Objects"))
					if objectsMap.DataType != "any" {
						yks.Throw(inter.CurrentFileName, "Field Objects must be a table with datatype 'any', not '%s'", x, y, objectsMap.DataType)
					}

					shaderProgramObj := sigmaMustAssert[*yks.StructObject](workspaceObj.Get("ShaderProgram"))
					ok, reason = shaderProgramObj.CheckFormat([2]string{"program", "u32"})
					if !ok {
						yks.Throw(inter.CurrentFileName, reason, x, y)
					}

					shaderProgram := ShaderProgram{
						program: sigmaMustAssert[uint32](shaderProgramObj.Get("program")),

						savedLocations: make(map[string]int32),
					}

					enable := make([]uint32, enableMap.Len())
					disable := make([]uint32, disableMap.Len())

					i := 0
					for _, cell := range enableMap.AllFromFront() {
						enableFlag := cell.Get().(uint32)

						enable[i] = enableFlag
						i++
					}

					i = 0
					for _, cell := range disableMap.AllFromFront() {
						disableFlag := cell.Get().(uint32)

						disable[i] = disableFlag
						i++
					}

					objects := make(map[string]Object)

					for k := range objectsMap.AllFromFront() {
						name, ok := k.(string)
						if !ok {
							continue
						}

						object, ok := mainGame.Objects[name]
						if !ok {
							warn(fmt.Sprintf("Object '%s' doesn't exist in game's storage. Object was skipped", name))
							continue
						}

						object.SyncWithScript()

						objects[name] = object
					}

					workspace := &Workspace{
						Name: name,

						Game: mainGame,

						UseShadowMaps: useShadowMaps,

						Enable:  enable,
						Disable: disable,

						DepthMask: depthMask,
						DepthFunc: depthFunc,

						Factor: factor,
						Units:  units,

						CullFace: cullFace,

						ShaderProgram: shaderProgram,

						Objects: objects,

						ScriptWorkspace: workspaceObj,
					}

					mainGame.AddWorkspace(workspace)

					return []any{}
				}),
			},
			{
				Identifier: "AddShadowMap",
				DataType:   "func",
				Method:     true,

				//* AddShadowMap
				Func: yks.NewFTemp("AddShadowMap", func(v ...any) []any {
					yks.ArgsCheck(v, 1, 1, "ShadowMap")

					x, y := v[0].(int), v[1].(int)
					inter := v[2].(*yks.Interpreter)

					v = v[yks.BUILTIN_SPECIALS:]

					shadowMapObj := v[0].(*yks.StructObject)

					ok, reason := shadowMapObj.CheckFormat(
						[2]string{"UniformName", "string"},

						[2]string{"ShaderProgram", "ShaderProgram"},

						[2]string{"Layers", "i32"},
						[2]string{"Resolution", "i32"},

						[2]string{"TextureUnit", "u32"},
						[2]string{"DepthMap", "u32"},

						[2]string{"Target", "u32"},

						[2]string{"FBO", "u32"},
					)
					if !ok {
						yks.Throw(inter.CurrentFileName, reason, x, y)
					}

					uniformName := sigmaMustAssert[string](shadowMapObj.Get("UniformName"))

					layers := sigmaMustAssert[int32](shadowMapObj.Get("Layers"))

					resolution := sigmaMustAssert[int32](shadowMapObj.Get("Resolution"))
					textureUnit := sigmaMustAssert[uint32](shadowMapObj.Get("TextureUnit"))

					depthMap := sigmaMustAssert[uint32](shadowMapObj.Get("DepthMap"))
					target := sigmaMustAssert[uint32](shadowMapObj.Get("Target"))

					fbo := sigmaMustAssert[uint32](shadowMapObj.Get("FBO"))

					shaderProgramObj := sigmaMustAssert[*yks.StructObject](shadowMapObj.Get("ShaderProgram"))
					ok, reason = shaderProgramObj.CheckFormat([2]string{"program", "u32"})
					if !ok {
						yks.Throw(inter.CurrentFileName, reason, x, y)
					}

					shaderProgram := ShaderProgram{
						program: sigmaMustAssert[uint32](shaderProgramObj.Get("program")),

						savedLocations: make(map[string]int32),
					}

					shadowMap := ShadowMap{
						UniformName: uniformName,

						ShaderProgram: shaderProgram,

						Layers: layers,

						Resolution: resolution,

						Target: target,

						FBO:      fbo,
						DepthMap: depthMap,

						TextureUnit: textureUnit,
					}

					mainGame.AddShadowMap(shadowMap)

					return []any{}
				}),
			},

			{
				Identifier: "ListenInput",
				DataType:   "func",
				Method:     true,

				//* ListenInput
				Func: yks.NewFTemp("ListenInput", func(v ...any) []any {
					yks.ArgsCheck(v, 2, 2, "i64", "bool")

					//x, y := v[0].(int), v[1].(int)
					//inter := v[2].(*yks.Interpreter)

					v = v[yks.BUILTIN_SPECIALS:]

					key := v[0].(int64)
					isMouse := v[1].(bool)

					if !isMouse {
						mainGame.ListenKeys = append(mainGame.ListenKeys, key)
					} else {
						mainGame.ListenButtons = append(mainGame.ListenButtons, key)
					}

					return []any{}
				}),
			},

			{
				Identifier: "GetCamera",
				DataType:   "func",
				Method:     true,

				//* GetCamera
				Func: yks.NewFTemp("GetCamera", func(v ...any) []any {
					x, y := v[0].(int), v[1].(int)
					inter := v[2].(*yks.Interpreter)

					v = v[yks.BUILTIN_SPECIALS:]

					camera := mainGame.Camera
					if camera.ScriptCamera != nil {
						return []any{camera.ScriptCamera}
					}

					structure, ok := sigmaAssertB[*yks.Structure](inter.CurrentScope.Get("Camera"))
					if !ok {
						return []any{nil}
					}

					vec3Structure, ok := sigmaAssertB[*yks.Structure](inter.CurrentScope.Get("Vec3"))
					if !ok {
						return []any{nil}
					}

					var cameraObj *yks.StructObject
					{

						positionObj := makeStructObjectFromStructure(vec3Structure, map[string]*yks.Field{
							"X": {
								Identifier: "X",
								DataType:   "f32",

								Value: yks.CLPTR(inter.CurrentScope, "f32", float32(0), x, y),
							},
							"Y": {
								Identifier: "Y",
								DataType:   "f32",

								Value: yks.CLPTR(inter.CurrentScope, "f32", float32(0), x, y),
							},
							"Z": {
								Identifier: "Z",
								DataType:   "f32",

								Value: yks.CLPTR(inter.CurrentScope, "f32", float32(0), x, y),
							},
						})

						frontObj := makeStructObjectFromStructure(vec3Structure, map[string]*yks.Field{
							"X": {
								Identifier: "X",
								DataType:   "f32",

								Value: yks.CLPTR(inter.CurrentScope, "f32", float32(0), x, y),
							},
							"Y": {
								Identifier: "Y",
								DataType:   "f32",

								Value: yks.CLPTR(inter.CurrentScope, "f32", float32(0), x, y),
							},
							"Z": {
								Identifier: "Z",
								DataType:   "f32",

								Value: yks.CLPTR(inter.CurrentScope, "f32", float32(-1), x, y),
							},
						})

						upObj := makeStructObjectFromStructure(vec3Structure, map[string]*yks.Field{
							"X": {
								Identifier: "X",
								DataType:   "f32",

								Value: yks.CLPTR(inter.CurrentScope, "f32", float32(0), x, y),
							},
							"Y": {
								Identifier: "Y",
								DataType:   "f32",

								Value: yks.CLPTR(inter.CurrentScope, "f32", float32(1), x, y),
							},
							"Z": {
								Identifier: "Z",
								DataType:   "f32",

								Value: yks.CLPTR(inter.CurrentScope, "f32", float32(0), x, y),
							},
						})

						cameraUpObj := makeStructObjectFromStructure(vec3Structure, map[string]*yks.Field{
							"X": {
								Identifier: "X",
								DataType:   "f32",

								Value: yks.CLPTR(inter.CurrentScope, "f32", float32(0), x, y),
							},
							"Y": {
								Identifier: "Y",
								DataType:   "f32",

								Value: yks.CLPTR(inter.CurrentScope, "f32", float32(1), x, y),
							},
							"Z": {
								Identifier: "Z",
								DataType:   "f32",

								Value: yks.CLPTR(inter.CurrentScope, "f32", float32(0), x, y),
							},
						})

						cameraRightObj := makeStructObjectFromStructure(vec3Structure, map[string]*yks.Field{
							"X": {
								Identifier: "X",
								DataType:   "f32",

								Value: yks.CLPTR(inter.CurrentScope, "f32", float32(0), x, y),
							},
							"Y": {
								Identifier: "Y",
								DataType:   "f32",

								Value: yks.CLPTR(inter.CurrentScope, "f32", float32(0), x, y),
							},
							"Z": {
								Identifier: "Z",
								DataType:   "f32",

								Value: yks.CLPTR(inter.CurrentScope, "f32", float32(0), x, y),
							},
						})

						cameraObj = makeStructObjectFromStructure(structure, map[string]*yks.Field{
							"Position": {
								Identifier: "Position",
								DataType:   "Vec3",

								Value: yks.CLPTR(inter.CurrentScope, "Vec3", positionObj, x, y),
							},
							"Front": {
								Identifier: "Front",
								DataType:   "Vec3",

								Value: yks.CLPTR(inter.CurrentScope, "Vec3", frontObj, x, y),
							},

							"Up": {
								Identifier: "Up",
								DataType:   "Vec3",

								Value: yks.CLPTR(inter.CurrentScope, "Vec3", upObj, x, y),
							},

							"CameraRight": {
								Identifier: "CameraRight",
								DataType:   "Vec3",

								Value: yks.CLPTR(inter.CurrentScope, "Vec3", cameraRightObj, x, y),
							},
							"CameraUp": {
								Identifier: "CameraUp",
								DataType:   "Vec3",

								Value: yks.CLPTR(inter.CurrentScope, "Vec3", cameraUpObj, x, y),
							},
						})
					}

					camera.ScriptCamera = cameraObj

					return []any{cameraObj}
				}),
			},
		},
	}

	builtinVals = []yks.BuiltinVal{
		{Key: "LoadMesh", Val: func(v ...any) []any {
			yks.ArgsCheck(v, 1, 1, "string")

			x, y := v[0].(int), v[1].(int)
			inter := v[2].(*yks.Interpreter)

			v = v[yks.BUILTIN_SPECIALS:]

			name := v[0].(string)

			mesh := newMeshFromFile(name, gl.STATIC_DRAW, false, Attribute{0, 3, gl.FLOAT, false, vertexStride, 0},
				Attribute{1, 3, gl.FLOAT, false, vertexStride, uintptr(3 * 4)},
				Attribute{2, 2, gl.FLOAT, false, vertexStride, uintptr(6 * 4)},
				Attribute{3, 4, gl.UNSIGNED_BYTE, false, vertexStride, uintptr(8 * 4)},
				Attribute{4, 4, gl.FLOAT, false, vertexStride, uintptr(8*4 + 4)},
				Attribute{5, 4, gl.FLOAT, false, vertexStride, uintptr(12*4 + 4)})

			structure, ok := sigmaAssertB[*yks.Structure](inter.CurrentScope.Get("Mesh"))
			if !ok {
				return []any{nil}
			}

			meshObj := makeStructObjectFromStructure(structure, map[string]*yks.Field{
				"Name": {
					Identifier: "Name",

					DataType: "string",

					Value: yks.CLPTR(inter.CurrentScope, "string", name, x, y),
				},
			})

			mesh.ScriptMesh = meshObj

			mainGame.AddMesh(mesh)

			return []any{meshObj}
		}},

		{Key: "LoadShader", Val: func(v ...any) []any {
			yks.ArgsCheck(v, 2, 2, "string", "u32")

			x, y := v[0].(int), v[1].(int)
			inter := v[2].(*yks.Interpreter)

			v = v[yks.BUILTIN_SPECIALS:]

			name := v[0].(string)

			stype := v[1].(uint32)

			shader, ok := newShaderFromFile(name, stype)
			if !ok {
				return []any{nil, false}
			}
			structure, ok := sigmaAssertB[*yks.Structure](inter.CurrentScope.Get("Shader"))
			if !ok {
				return []any{nil}
			}

			shaderObj := makeStructObjectFromStructure(structure, map[string]*yks.Field{
				"Source": {
					Identifier: "Source",
					DataType:   "string",

					Value: yks.CLPTR(inter.CurrentScope, "string", shader.Source, x, y),
				},
				"Type": {
					Identifier: "Type",
					DataType:   "u32",

					Value: yks.CLPTR(inter.CurrentScope, "u32", shader.Type, x, y),
				},
				"shader": {
					Identifier: "shader",
					DataType:   "u32",

					Value: yks.CLPTR(inter.CurrentScope, "u32", shader.shader, x, y),
				},
			})

			return []any{shaderObj}
		}},

		{Key: "MakeShaderProgram", Val: func(v ...any) []any {
			yks.ArgsCheck(v, 1, 1, "table")

			x, y := v[0].(int), v[1].(int)
			inter := v[2].(*yks.Interpreter)

			v = v[yks.BUILTIN_SPECIALS:]

			shadersMap := v[0].(*yks.Map)

			structure, ok := sigmaAssertB[*yks.Structure](inter.CurrentScope.Get("ShaderProgram"))
			if !ok {
				return []any{nil}
			}

			shaders := make([]Shader, shadersMap.Len())

			i := -1
			for _, shaderCell := range shadersMap.AllFromFront() {
				i++
				shaderObj, ok := shaderCell.Get().(*yks.StructObject)
				if !ok {
					return []any{nil}
				}

				ok, reason := shaderObj.CheckFormat([2]string{"Source", "string"}, [2]string{"Type", "u32"}, [2]string{"shader", "u32"})
				if !ok {
					yks.Throw(inter.CurrentFileName, reason, x, y)
				}

				shader := Shader{
					Source: sigmaMustAssert[string](shaderObj.Get("Source")),

					Type:   sigmaMustAssert[uint32](shaderObj.Get("Type")),
					shader: sigmaMustAssert[uint32](shaderObj.Get("shader")),
				}

				shaders[i] = shader
			}

			program, ok := newShaderProgram()
			if !ok {
				return []any{nil}
			}

			program.AttachShaders(true, shaders...)
			program.Link()

			programObj := makeStructObjectFromStructure(structure, map[string]*yks.Field{
				"program": {
					Identifier: "program",
					DataType:   "u32",

					Value: yks.CLPTR(inter.CurrentScope, "u32", program.program, x, y),
				},
			})

			return []any{programObj}
		}},

		{Key: "NewShadowMap", Val: func(v ...any) []any {
			yks.ArgsCheck(v, 5, 5, "ShaderProgram", "string", "i32", "i32", "u32")

			x, y := v[0].(int), v[1].(int)
			inter := v[2].(*yks.Interpreter)

			v = v[yks.BUILTIN_SPECIALS:]

			shaderProgramObj := v[0].(*yks.StructObject)

			ok, reason := shaderProgramObj.CheckFormat([2]string{"program", "u32"})
			if !ok {
				yks.Throw(inter.CurrentFileName, reason, x, y)
			}

			shaderProgram := ShaderProgram{
				program: sigmaMustAssert[uint32](shaderProgramObj.Get("program")),
			}

			uniformName := v[1].(string)
			resolution := v[2].(int32)
			layers := v[3].(int32)

			texUnit := v[4].(uint32)

			structure, ok := sigmaAssertB[*yks.Structure](inter.CurrentScope.Get("ShadowMap"))
			if !ok {
				return []any{nil}
			}

			textureUnit := gl.TEXTURE10 + texUnit

			shadowMap := newShadowMap(shaderProgram, uniformName, resolution, layers, textureUnit, mgl32.Vec4{1, 1, 1, 1})

			shadowMapObj := makeStructObjectFromStructure(structure, map[string]*yks.Field{
				"UniformName": {
					Identifier: "UniformName",
					DataType:   "string",

					Value: yks.CLPTR(inter.CurrentScope, "string", shadowMap.UniformName, x, y),
				},
				"ShaderProgram": {
					Identifier: "ShaderProgram",
					DataType:   "ShaderProgram",

					Value: yks.CLPTR(inter.CurrentScope, "ShaderProgram", shaderProgramObj, x, y),
				},
				"Layers": {
					Identifier: "Layers",
					DataType:   "i32",

					Value: yks.CLPTR(inter.CurrentScope, "i32", shadowMap.Layers, x, y),
				},
				"Resolution": {
					Identifier: "Resolution",
					DataType:   "i32",

					Value: yks.CLPTR(inter.CurrentScope, "i32", shadowMap.Resolution, x, y),
				},
				"TextureUnit": {
					Identifier: "TextureUnit",
					DataType:   "u32",

					Value: yks.CLPTR(inter.CurrentScope, "u32", shadowMap.TextureUnit, x, y),
				},
				"DepthMap": {
					Identifier: "DepthMap",
					DataType:   "u32",

					Value: yks.CLPTR(inter.CurrentScope, "u32", shadowMap.DepthMap, x, y),
				},
				"FBO": {
					Identifier: "FBO",
					DataType:   "u32",

					Value: yks.CLPTR(inter.CurrentScope, "u32", shadowMap.FBO, x, y),
				},
				"Target": {
					Identifier: "Target",
					DataType:   "u32",

					Value: yks.CLPTR(inter.CurrentScope, "u32", shadowMap.Target, x, y),
				},
			})

			return []any{shadowMapObj}
		}},

		{Key: "NewShadowCubeMap", Val: func(v ...any) []any {
			yks.ArgsCheck(v, 5, 5, "ShaderProgram", "string", "i32", "i32", "u32")

			x, y := v[0].(int), v[1].(int)
			inter := v[2].(*yks.Interpreter)

			v = v[yks.BUILTIN_SPECIALS:]

			shaderProgramObj := v[0].(*yks.StructObject)

			ok, reason := shaderProgramObj.CheckFormat([2]string{"program", "u32"})
			if !ok {
				yks.Throw(inter.CurrentFileName, reason, x, y)
			}

			shaderProgram := ShaderProgram{
				program: sigmaMustAssert[uint32](shaderProgramObj.Get("program")),
			}

			uniformName := v[1].(string)
			resolution := v[2].(int32)
			layers := v[3].(int32)

			texUnit := v[4].(uint32)

			structure, ok := sigmaAssertB[*yks.Structure](inter.CurrentScope.Get("ShadowMap"))
			if !ok {
				return []any{nil}
			}

			textureUnit := gl.TEXTURE10 + texUnit

			shadowMap := newShadowMapCubeMap(shaderProgram, uniformName, resolution, layers, textureUnit)

			shadowMapObj := makeStructObjectFromStructure(structure, map[string]*yks.Field{
				"UniformName": {
					Identifier: "UniformName",
					DataType:   "string",

					Value: yks.CLPTR(inter.CurrentScope, "string", shadowMap.UniformName, x, y),
				},
				"ShaderProgram": {
					Identifier: "ShaderProgram",
					DataType:   "ShaderProgram",

					Value: yks.CLPTR(inter.CurrentScope, "ShaderProgram", shaderProgramObj, x, y),
				},
				"Layers": {
					Identifier: "Layers",
					DataType:   "i32",

					Value: yks.CLPTR(inter.CurrentScope, "i32", shadowMap.Layers, x, y),
				},
				"Resolution": {
					Identifier: "Resolution",
					DataType:   "i32",

					Value: yks.CLPTR(inter.CurrentScope, "i32", shadowMap.Resolution, x, y),
				},
				"TextureUnit": {
					Identifier: "TextureUnit",
					DataType:   "u32",

					Value: yks.CLPTR(inter.CurrentScope, "u32", shadowMap.TextureUnit, x, y),
				},
				"DepthMap": {
					Identifier: "DepthMap",
					DataType:   "u32",

					Value: yks.CLPTR(inter.CurrentScope, "u32", shadowMap.DepthMap, x, y),
				},
				"FBO": {
					Identifier: "FBO",
					DataType:   "u32",

					Value: yks.CLPTR(inter.CurrentScope, "u32", shadowMap.FBO, x, y),
				},
				"Target": {
					Identifier: "Target",
					DataType:   "u32",

					Value: yks.CLPTR(inter.CurrentScope, "u32", shadowMap.Target, x, y),
				},
			})

			return []any{shadowMapObj}
		}},

		{Key: "NewMeshObject", Val: func(v ...any) []any {
			yks.ArgsCheck(v, 2, 2, "string", "Mesh")

			x, y := v[0].(int), v[1].(int)
			inter := v[2].(*yks.Interpreter)

			v = v[yks.BUILTIN_SPECIALS:]

			name := v[0].(string)

			meshObj := v[1].(*yks.StructObject)
			ok, reason := meshObj.CheckFormat([2]string{"Name", "string"})
			if !ok {
				yks.Throw(inter.CurrentFileName, reason, x, y)
			}

			structure, ok := sigmaAssertB[*yks.Structure](inter.CurrentScope.Get("MeshObject"))
			if !ok {
				return []any{nil}
			}

			mesh := mainGame.GetMesh(sigmaMustAssert[string](meshObj.Get("Name")))

			meshObject := newMeshObject(mesh, mgl32.Vec3{}, mgl32.Vec3{1, 1, 1}, mgl32.QuatIdent())

			animationObjMap := &yks.Map{
				OrderedMap: orderedmap.NewOrderedMap[any, *yks.Cell](),
				DataType:   "Animation",
				Pointers:   []any{},
				Layout:     []string{},
				Mem:        []byte{},
			}

			animationStructure, ok := sigmaAssertB[*yks.Structure](inter.CurrentScope.Get("Animation"))
			if !ok {
				return []any{nil}
			}

			vec3Structure, ok := sigmaAssertB[*yks.Structure](inter.CurrentScope.Get("Vec3"))
			if !ok {
				return []any{nil}
			}

			quatStructure, ok := sigmaAssertB[*yks.Structure](inter.CurrentScope.Get("Quat"))
			if !ok {
				return []any{nil}
			}

			var positionObj, scaleObj, rotationVec3Obj, rotationObj *yks.StructObject
			{
				positionObj = makeStructObjectFromStructure(vec3Structure, map[string]*yks.Field{
					"X": {
						Identifier: "X",
						DataType:   "f32",

						Value: yks.CLPTR(inter.CurrentScope, "f32", float32(0), x, y),
					},
					"Y": {
						Identifier: "Y",
						DataType:   "f32",

						Value: yks.CLPTR(inter.CurrentScope, "f32", float32(0), x, y),
					},
					"Z": {
						Identifier: "Z",
						DataType:   "f32",

						Value: yks.CLPTR(inter.CurrentScope, "f32", float32(0), x, y),
					},
				})

				scaleObj = makeStructObjectFromStructure(vec3Structure, map[string]*yks.Field{
					"X": {
						Identifier: "X",
						DataType:   "f32",

						Value: yks.CLPTR(inter.CurrentScope, "f32", float32(1), x, y),
					},
					"Y": {
						Identifier: "Y",
						DataType:   "f32",

						Value: yks.CLPTR(inter.CurrentScope, "f32", float32(1), x, y),
					},
					"Z": {
						Identifier: "Z",
						DataType:   "f32",

						Value: yks.CLPTR(inter.CurrentScope, "f32", float32(1), x, y),
					},
				})

				rotationVec3Obj = makeStructObjectFromStructure(vec3Structure, map[string]*yks.Field{
					"X": {
						Identifier: "X",
						DataType:   "f32",

						Value: yks.CLPTR(inter.CurrentScope, "f32", float32(0), x, y),
					},
					"Y": {
						Identifier: "Y",
						DataType:   "f32",

						Value: yks.CLPTR(inter.CurrentScope, "f32", float32(0), x, y),
					},
					"Z": {
						Identifier: "Z",
						DataType:   "f32",

						Value: yks.CLPTR(inter.CurrentScope, "f32", float32(0), x, y),
					},
				})

				rotationObj = makeStructObjectFromStructure(quatStructure, map[string]*yks.Field{
					"V": {
						Identifier: "V",
						DataType:   "Vec3",

						Value: yks.CLPTR(inter.CurrentScope, "Vec3", rotationVec3Obj, x, y),
					},
					"W": {
						Identifier: "W",
						DataType:   "f32",

						Value: yks.CLPTR(inter.CurrentScope, "f32", float32(1), x, y),
					},
				})
			}

			meshObjectObj := makeStructObjectFromStructure(structure, map[string]*yks.Field{
				"Name": {
					Identifier: "Name",
					DataType:   "string",

					Value: yks.CLPTR(inter.CurrentScope, "string", name, x, y),
				},
				"Mesh": {
					Identifier: "Mesh",
					DataType:   "Mesh",

					Value: yks.CLPTR(inter.CurrentScope, "Mesh", meshObj, x, y),
				},
				"Position": {
					Identifier: "Position",
					DataType:   "Vec3",

					Value: yks.CLPTR(inter.CurrentScope, "Vec3", positionObj, x, y),
				},
				"Scale": {
					Identifier: "Scale",
					DataType:   "Vec3",

					Value: yks.CLPTR(inter.CurrentScope, "Vec3", scaleObj, x, y),
				},
				"Rotation": {
					Identifier: "Rotation",
					DataType:   "Quat",

					Value: yks.CLPTR(inter.CurrentScope, "Quat", rotationObj, x, y),
				},
				"Animations": {
					Identifier: "Animations",
					DataType:   "table",

					Value: yks.CLPTR(inter.CurrentScope, "table", animationObjMap, x, y),
				},
			})

			for i, animation := range meshObject.Animations {
				animationObj := makeStructObjectFromStructure(animationStructure, map[string]*yks.Field{
					"Mesh": {
						Identifier: "Mesh",
						DataType:   "Mesh",

						Value: yks.CLPTR(inter.CurrentScope, "Mesh", meshObj, x, y),
					},
					"MeshObject": {
						Identifier: "MeshObject",
						DataType:   "MeshObject",

						Value: yks.CLPTR(inter.CurrentScope, "MeshObject", meshObjectObj, x, y),
					},

					"TimeMarker": {
						Identifier: "TimeMarker",
						DataType:   "f32",

						Value: yks.CLPTR(inter.CurrentScope, "f32", animation.TimeMarker, x, y),
					},
					"IsPlaying": {
						Identifier: "IsPlaying",
						DataType:   "bool",

						Value: yks.CLPTR(inter.CurrentScope, "bool", animation.IsPlaying, x, y),
					},
					"Looped": {
						Identifier: "Looped",
						DataType:   "bool",

						Value: yks.CLPTR(inter.CurrentScope, "bool", animation.Looped, x, y),
					},
				})

				animation.ScriptAnimation = animationObj

				animationObjMap.Set(int64(i), yks.CLPTR(inter.CurrentScope, "Animation", animationObj, x, y))
			}

			meshObject.ScriptMeshObject = meshObjectObj

			mainGame.AddObject(name, meshObject)

			return []any{meshObjectObj}
		}},

		{Key: "OS_NAME", Val: func(v ...any) []any {
			return []any{runtime.GOOS}
		}},

		{Key: "print", Val: func(v ...any) []any {
			fmt.Println(yks.Format(false, v[yks.BUILTIN_SPECIALS:]...))
			return nil
		}},

		{Key: "delete", Val: func(v ...any) []any {
			yks.ArgsCheck(v, 2, 2, "table", "any")

			v = v[yks.BUILTIN_SPECIALS:]

			table := v[0].(*orderedmap.OrderedMap[any, *yks.Cell])
			key := v[1]

			table.Delete(key)
			return nil
		}},

		{Key: "sleep", Val: func(v ...any) []any {
			x, y := v[0].(int), v[1].(int)
			inter := v[2].(*yks.Interpreter)

			if len(v) == 0 {
				Throw(inter.CurrentFileName, "Function must have one argument.", x, y)
			}

			v = v[yks.BUILTIN_SPECIALS:]

			switch t := v[0].(type) {
			case float64, float32:
				time.Sleep(time.Duration(yks.MustNTOF64(t) * float64(time.Second)))
			case int64, int32, int16, int8:
				time.Sleep(time.Duration(yks.ToInt64(t) * int64(time.Millisecond)))
			case uint64, uint32, uint16, uint8:
				time.Sleep(time.Duration(yks.ToUint64(t) * uint64(time.Millisecond)))
			default:
				Throw(inter.CurrentFileName, "Time value must be a number.", x, y)
			}
			return nil
		}},

		{Key: "throw", Val: func(v ...any) []any {
			x, y := v[0].(int), v[1].(int)
			inter := v[2].(*yks.Interpreter)

			v = v[yks.BUILTIN_SPECIALS:]
			if len(v) <= 0 {
				Throw(inter.CurrentFileName, "Function requires one or more arguments.", x, y)
			}

			Throw(inter.CurrentFileName, yks.Format(false, v...), x, y)
			return nil
		}},

		{Key: "len", Val: func(v ...any) []any {
			yks.ArgsCheck(v, 1, 1, "any")
			x, y := v[0].(int), v[1].(int)
			inter := v[2].(*yks.Interpreter)

			v = v[yks.BUILTIN_SPECIALS:]

			a := v[0]
			switch a := a.(type) {
			case *yks.Map:
				return []any{int64(a.Len())}
			case string:
				return []any{int64(len(a))}
			case *yks.StructObject:
				layout := a.Layout()
				if len(layout) == 0 {
					return []any{int64(0)}
				}

				lastFieldLayout := layout[len(layout)-1]

				return []any{int64(lastFieldLayout.Offset + lastFieldLayout.Size)}
			default:
				Throw(inter.CurrentFileName, "Cannot get lenght of non-string, non-table or non-instance value.", x, y)
			}
			return nil
		}},

		{Key: "sizeof", Val: func(v ...any) []any {
			yks.ArgsCheck(v, 1, 1, "any")

			v = v[yks.BUILTIN_SPECIALS:]

			a := v[0]
			switch v := a.(type) {
			case *yks.Map:
				a = v.Mem
			}

			return []any{unsafe.Sizeof(a)}
		}},

		{Key: "time", Val: func(v ...any) []any {
			return []any{time.Now().UnixMilli()}
		}},
		{Key: "strformat", Val: func(v ...any) []any {
			return []any{yks.Format(false, v[yks.BUILTIN_SPECIALS:]...)}
		}},
		{Key: "gettype", Val: func(v ...any) []any {
			yks.ArgsCheck(v, 1, 1, "any")

			v = v[yks.BUILTIN_SPECIALS:]

			return []any{yks.GetValueType(v[0])}
		}},
		{Key: "numformat", Val: func(v ...any) []any {
			yks.ArgsCheck(v, 2, 2, "string", "bool")
			x, y := v[0].(int), v[1].(int)
			inter := v[2].(*yks.Interpreter)

			v = v[yks.BUILTIN_SPECIALS:]

			str := v[0].(string)
			isint := v[1].(bool)

			if !isint {
				n, err := strconv.ParseFloat(str, 64)
				switch err {
				case strconv.ErrSyntax:
					Throw(inter.CurrentFileName, "Syntax error while trying to parse number value.", x, y)
				case strconv.ErrRange:
					Throw(inter.CurrentFileName, "Number value is out of range.", x, y)
				}
				return []any{n}
			} else {
				n, err := strconv.ParseInt(str, 0, 64)
				switch err {
				case strconv.ErrSyntax:
					Throw(inter.CurrentFileName, "Syntax error while trying to parse number value.", x, y)
				case strconv.ErrRange:
					Throw(inter.CurrentFileName, "Number value is out of range.", x, y)
				}

				return []any{n}
			}
		}},

		{Key: "string", Val: func(v ...any) []any {
			yks.ArgsCheck(v, 1, 1, "table")

			v = v[yks.BUILTIN_SPECIALS:]

			b := v[0].(*yks.Map)
			bstring := []byte{}

		APPEND:
			for _, v := range b.AllFromFront() {
				switch v := v.Get().(type) {
				case int64, int32, int16, int8, uint8, uint16, uint32, uint64:
					charByte := yks.ToUint(yks.ToUint64(v), 8).(byte)

					bstring = append(bstring, charByte)
				default:
					log.Println("Unknown datatype lol the developer is such a shitcoder")
					break APPEND
				}
			}

			return []any{string(bstring)}
		}},

		{Key: "unicodetostr", Val: func(v ...any) []any {
			yks.ArgsCheck(v, 1, 1, "uint")

			v = v[yks.BUILTIN_SPECIALS:]

			r := rune(yks.ToUint64(v[0]))

			return []any{string(r)}
		}},

		{Key: "make", Val: func(v ...any) []any {
			yks.ArgsCheck(v, 3, 3, "int", "string", "any")

			x, y := v[0].(int), v[1].(int)
			inter := v[2].(*yks.Interpreter)

			v = v[yks.BUILTIN_SPECIALS:]

			length := int(yks.ToInt64(v[0]))
			dataType := v[1].(string)
			defaultValue := v[2]

			if length < 0 {
				length = 0
			}

			m := &yks.Map{
				OrderedMap: orderedmap.NewOrderedMap[any, *yks.Cell](),

				DataType: dataType,

				Pointers: []any{},
				Layout:   []string{},
				Mem:      []byte{},
			}

			for i := 0; i < length; i++ {
				m.Set(int64(i), yks.CLPTR(inter.CurrentScope, dataType, defaultValue, x, y))
			}
			m.ToMemory()

			return []any{m}
		}},

		{Key: "cstring", Val: func(v ...any) []any {
			yks.ArgsCheck(v, 1, 1, "string")

			v = v[yks.BUILTIN_SPECIALS:]

			str := v[0].(string)

			slicePtr, err := syscall.BytePtrFromString(str)
			if err == nil {
				err = errors.New("Successfull")
			}

			return []any{
				uintptr(unsafe.Pointer(slicePtr)), err,
			}
		}},

		{Key: "bytes", Val: func(v ...any) []any {
			yks.ArgsCheck(v, 1, 1, "string")

			x, y := v[0].(int), v[1].(int)
			inter := v[2].(*yks.Interpreter)

			v = v[yks.BUILTIN_SPECIALS:]

			str := v[0].(string)

			slice, err := syscall.ByteSliceFromString(str)
			handle(err)

			m := &yks.Map{
				OrderedMap: orderedmap.NewOrderedMap[any, *yks.Cell](),
				DataType:   "u8",
				Pointers:   []any{},
				Layout:     []string{},
				Mem:        []byte{},
			}

			for i, v := range slice {
				m.Set(int64(i), yks.CLPTR(inter.CurrentScope, "u8", uint8(v), x, y))
			}
			m.ToMemory()

			return []any{
				m,
			}
		}},

		{Key: "cos32", Val: func(v ...any) []any {
			yks.ArgsCheck(v, 1, 1, "f32")

			v = v[yks.BUILTIN_SPECIALS:]

			f := v[0].(float32)

			return []any{float32(math.Cos(float64(f)))}
		}},

		{Key: "cos", Val: func(v ...any) []any {
			yks.ArgsCheck(v, 1, 1, "f64")

			v = v[yks.BUILTIN_SPECIALS:]

			f := v[0].(float64)

			return []any{math.Cos(f)}
		}},

		{Key: "sin32", Val: func(v ...any) []any {
			yks.ArgsCheck(v, 1, 1, "f32")

			v = v[yks.BUILTIN_SPECIALS:]

			f := v[0].(float32)

			return []any{float32(math.Sin(float64(f)))}
		}},

		{Key: "sin", Val: func(v ...any) []any {
			yks.ArgsCheck(v, 1, 1, "f64")

			v = v[yks.BUILTIN_SPECIALS:]

			f := v[0].(float64)

			return []any{math.Sin(f)}
		}},

		{Key: "tan32", Val: func(v ...any) []any {
			yks.ArgsCheck(v, 1, 1, "f32")

			v = v[yks.BUILTIN_SPECIALS:]

			f := v[0].(float32)

			return []any{float32(math.Tan(float64(f)))}
		}},

		{Key: "tan", Val: func(v ...any) []any {
			yks.ArgsCheck(v, 1, 1, "f64")

			v = v[yks.BUILTIN_SPECIALS:]

			f := v[0].(float64)

			return []any{math.Tan(f)}
		}},

		{Key: "Game", Val: gameYKSStructure},

		{Key: "game", Val: makeStructObjectFromStructure(gameYKSStructure, map[string]*yks.Field{})},
	}
)
