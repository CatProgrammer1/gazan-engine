package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

func newShader(source string, stype uint32) (Shader, bool) {
	if !strings.HasSuffix(source, "\x00") {
		source += "\x00"
	}

	shader := Shader{
		Source:  source,
		Type:    stype,
		deleted: false,
	}

	sourcePtr, free := gl.Strs(source)
	defer free()

	shader.shader = gl.CreateShader(stype)
	gl.ShaderSource(shader.shader, 1, sourcePtr, nil)
	gl.CompileShader(shader.shader)

	logBufferSize := int32(512)
	logInfo := make([]uint8, logBufferSize)

	var success int32
	gl.GetShaderiv(shader.shader, gl.COMPILE_STATUS, &success)
	if success != 1 {
		log.Println(source)
		gl.GetShaderInfoLog(shader.shader, logBufferSize, nil, &logInfo[0])
		log.Println(gl.GoStr(&logInfo[0]))
		return shader, false
	}

	return shader, true
}

func newShaderFromFile(path string, stype uint32) (Shader, bool) {
	source, err := os.ReadFile(path)
	handle(err)

	return newShader(string(append(source, 0)), stype)
}

type Shader struct {
	Source string
	Type   uint32
	shader uint32

	deleted bool
}

func (shader Shader) LogStatus() {
	if shader.deleted {
		panic("Shader no longer exists")
	}

	if shader.shader == 0 {
		panic("Invalid shader")
	}
}

func (shader Shader) Attach(program uint32, deleteWhenDone bool) {
	shader.LogStatus()

	gl.AttachShader(program, shader.shader)

	if deleteWhenDone {
		shader.Delete()
	}
}

func (shader Shader) Delete() {
	shader.LogStatus()

	gl.DeleteShader(shader.shader)
}

func newShaderProgram() (ShaderProgram, bool) {
	shaderProgram := ShaderProgram{
		program: gl.CreateProgram(),

		savedLocations: make(map[string]int32),
	}

	return shaderProgram, true
}

type ShaderProgram struct {
	program uint32

	savedLocations map[string]int32
}

func (shaderProgram ShaderProgram) AttachShaders(deleteWhenDone bool, shaders ...Shader) {
	for _, shader := range shaders {
		shader.Attach(shaderProgram.program, deleteWhenDone)
	}
}

func (shaderProgram ShaderProgram) Link() {
	gl.LinkProgram(shaderProgram.program)

	logBufferSize := int32(512)
	logInfo := make([]uint8, logBufferSize)

	var success int32

	gl.GetProgramiv(shaderProgram.program, gl.LINK_STATUS, &success)
	if success != 1 {
		gl.GetProgramInfoLog(shaderProgram.program, logBufferSize, nil, &logInfo[0])
		log.Println(gl.GoStr(&logInfo[0]))
	}
}

func (shaderProgram *ShaderProgram) GetUniformLocation(name string) int32 {
	location, ok := shaderProgram.savedLocations[name]
	if !ok {
		cString := append([]byte(name), 0)

		location = gl.GetUniformLocation(shaderProgram.program, (*uint8)(&cString[0]))

		shaderProgram.savedLocations[name] = location
	}

	return location
}

func (shaderProgram ShaderProgram) SetUniform(location int32, v any) {
	if location < 0 {
		return
	}

	switch v := v.(type) {
	case int32:
		gl.Uniform1i(location, v)
	case float32:
		gl.Uniform1f(location, v)
	case mgl32.Vec2:
		gl.Uniform2f(location, v[0], v[1])
	case mgl32.Vec3:
		gl.Uniform3f(location, v[0], v[1], v[2])
	case mgl32.Vec4:
		gl.Uniform4f(location, v[0], v[1], v[2], v[3])
	case mgl32.Mat4:
		gl.UniformMatrix4fv(location, 1, false, &v[0])
	case *CubeMap:
		v.Bind(location)
	default:
		fmt.Println("Unimplemented value type in auto SetUniform method")
	}
}

func (shaderProgram ShaderProgram) Use() {
	gl.UseProgram(shaderProgram.program)
}

func (shaderProgram ShaderProgram) Delete() {
	gl.DeleteProgram(shaderProgram.program)
}
