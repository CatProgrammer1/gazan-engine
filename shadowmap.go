package main

import (
	"log"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

func newShadowMap(shaderProgram ShaderProgram, uniformName string, resolution int32, layers int32, textureUnit uint32, borderColor mgl32.Vec4) ShadowMap {
	shadowMap := ShadowMap{
		ShaderProgram: shaderProgram,

		UniformName: uniformName,
	}

	var depthMapFBO uint32
	gl.GenFramebuffers(1, &depthMapFBO)

	var depthMap uint32
	gl.GenTextures(1, &depthMap)
	gl.BindTexture(gl.TEXTURE_2D_ARRAY, depthMap)
	gl.TexImage3D(gl.TEXTURE_2D_ARRAY, 0, gl.DEPTH_COMPONENT32F,
		resolution, resolution, layers, 0, gl.DEPTH_COMPONENT, gl.FLOAT, nil,
	)

	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_BORDER)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_BORDER)

	gl.TexParameterfv(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_BORDER_COLOR, &borderColor[0])

	gl.BindFramebuffer(gl.FRAMEBUFFER, depthMapFBO)
	gl.FramebufferTextureLayer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, depthMap, 0, 0)

	gl.DrawBuffer(gl.NONE)
	gl.ReadBuffer(gl.NONE)

	status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)
	if status != gl.FRAMEBUFFER_COMPLETE {
		log.Fatalln("HERE", status)
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	shadowMap.Resolution = resolution
	shadowMap.FBO = depthMapFBO
	shadowMap.DepthMap = depthMap
	shadowMap.TextureUnit = textureUnit
	shadowMap.Layers = layers
	shadowMap.Target = gl.TEXTURE_2D_ARRAY

	return shadowMap
}

func newShadowMapCubeMap(shaderProgram ShaderProgram, uniformName string, resolution int32, layers int32, textureUnit uint32) ShadowMap {
	shadowMap := ShadowMap{
		ShaderProgram: shaderProgram,

		UniformName: uniformName,
	}

	var depthMapFBO uint32
	gl.GenFramebuffers(1, &depthMapFBO)

	var depthMap uint32
	gl.GenTextures(1, &depthMap)

	gl.BindTexture(gl.TEXTURE_CUBE_MAP_ARRAY, depthMap)

	gl.TexStorage3D(gl.TEXTURE_CUBE_MAP_ARRAY, 1, gl.DEPTH_COMPONENT24, resolution, resolution, layers*6)

	gl.TexParameteri(gl.TEXTURE_CUBE_MAP_ARRAY, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP_ARRAY, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP_ARRAY, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP_ARRAY, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP_ARRAY, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)

	gl.BindFramebuffer(gl.FRAMEBUFFER, depthMapFBO)
	gl.FramebufferTexture(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, depthMap, 0)

	gl.DrawBuffer(gl.NONE)
	gl.ReadBuffer(gl.NONE)

	status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)
	if status != gl.FRAMEBUFFER_COMPLETE {
		log.Fatalf("Помилка Фреймбуфера: 0x%x", status)
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	shadowMap.Resolution = resolution
	shadowMap.FBO = depthMapFBO
	shadowMap.DepthMap = depthMap
	shadowMap.TextureUnit = textureUnit
	shadowMap.Layers = layers
	shadowMap.Target = gl.TEXTURE_CUBE_MAP_ARRAY

	return shadowMap
}

type ShadowMap struct {
	UniformName string

	//Depth ShaderProgram
	ShaderProgram ShaderProgram

	Layers int32

	Target uint32

	FBO         uint32
	DepthMap    uint32
	Resolution  int32
	TextureUnit uint32
}

func (sm ShadowMap) Bind() {
	shaderProgram := sm.ShaderProgram

	shaderProgram.Use()

	gl.BindFramebuffer(gl.FRAMEBUFFER, sm.FBO)
	gl.Viewport(0, 0, sm.Resolution, sm.Resolution)

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthMask(true)
	gl.DepthFunc(gl.LESS)
}

func (sm ShadowMap) SetUniform(shaderProgram ShaderProgram) {
	gl.Uniform1i(shaderProgram.GetUniformLocation(sm.UniformName), int32(sm.TextureUnit-gl.TEXTURE0))
}
