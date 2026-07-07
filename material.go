package main

import (
	"github.com/go-gl/gl/v4.3-core/gl"
)

func newMaterial(diffuse, normal, metalicRoughness *Texture2D, roughnessFactor, metallicFactor, shininess, reflectance, opacity float32) *Material {
	material := &Material{
		Diffuse:           diffuse,
		Shininess:         shininess,
		Reflectance:       reflectance,
		Normal:            normal,
		Opacity:           opacity,
		RoughnessFactor:   roughnessFactor,
		MetallicFactor:    metallicFactor,
		MetallicRoughness: metalicRoughness,

		savedLocations: make(map[uint32][]int32),
	}

	if normal == nil && defaultNormalTexture != nil {
		material.Normal = defaultNormalTexture
	}

	if metalicRoughness == nil && defaultMetallicRoughnessTexture != nil {
		material.MetallicRoughness = defaultMetallicRoughnessTexture
	}

	return material
}

type Material struct {
	Diffuse           *Texture2D
	Normal            *Texture2D
	MetallicRoughness *Texture2D
	RoughnessFactor   float32
	MetallicFactor    float32
	Shininess         float32
	Reflectance       float32
	Opacity           float32

	savedLocations map[uint32][]int32
}

func (material Material) Use(shaderProgram ShaderProgram, uniform string) {
	locations, ok := material.savedLocations[shaderProgram.program]

	if !ok {
		locations = []int32{
			shaderProgram.GetUniformLocation(uniform + ".diffuse"),
			shaderProgram.GetUniformLocation(uniform + ".normal"),

			shaderProgram.GetUniformLocation(uniform + ".shininess"),
			shaderProgram.GetUniformLocation(uniform + ".opacity"),

			shaderProgram.GetUniformLocation(uniform + ".metallicRoughness"),
			shaderProgram.GetUniformLocation(uniform + ".metallic"),
			shaderProgram.GetUniformLocation(uniform + ".roughness"),
		}

		/*material.Diffuse.SetSampleUnitLocation(shaderProgram, locations[0], 0)
		if material.Normal != nil {
			material.Normal.SetSampleUnitLocation(shaderProgram, locations[1], 1)
		}*/

		material.savedLocations[shaderProgram.program] = locations
	}

	material.MetallicRoughness.Bind()
	material.Normal.Bind()
	material.Diffuse.Bind()

	gl.Uniform1i(locations[0], 0)
	gl.Uniform1i(locations[1], 1)

	gl.Uniform1f(locations[2], material.Shininess)
	gl.Uniform1f(locations[3], material.Opacity)

	gl.Uniform1i(locations[4], 2)
	gl.Uniform1f(locations[5], material.MetallicFactor)
	gl.Uniform1f(locations[6], material.RoughnessFactor)
}
