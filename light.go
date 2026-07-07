package main

import (
	"math"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

func newLightSource(ltype int32, diffuse, position, direction mgl32.Vec3, maxDistance, innerCutOut, outerCutOut float32) *Light {
	light := &Light{
		Type:     ltype,
		Constant: 1, Linear: 0.7, Quadratic: 1.8,

		Position: position, Direction: direction, Diffuse: diffuse,

		MaxDistance: maxDistance,

		InnerCutOut: innerCutOut, OuterCutOut: outerCutOut,

		LightSpaceMatrix: [6]mgl32.Mat4{
			mgl32.Ident4(),
			mgl32.Ident4(),
			mgl32.Ident4(),
			mgl32.Ident4(),
			mgl32.Ident4(),
			mgl32.Ident4(),
		},

		savedLocations: make(map[uint32][]int32),

		OrthoLimits: []float32{-25, 25, -25, 25},
	}

	return light
}

func newPointLightSource(position, diffuse mgl32.Vec3, maxDistance float32) *Light {
	return newLightSource(1, diffuse, position, mgl32.Vec3{}, maxDistance, 0, 0)
}

func newSpotLightSource(position, direction, diffuse mgl32.Vec3, maxDistance, innerCutOut, outerCutOut float32) *Light {
	return newLightSource(2, diffuse, position, direction, maxDistance, innerCutOut, outerCutOut)
}

func newDirectionalLightSource(direction, diffuse mgl32.Vec3) *Light {
	return newLightSource(0, diffuse, mgl32.Vec3{}, direction, 0, 0, 0)
}

type Light struct {
	Type int32 //0 - directional, 1 - point

	Constant, Linear, Quadratic float32

	Position, Direction, Diffuse mgl32.Vec3

	MaxDistance float32

	InnerCutOut, OuterCutOut float32

	LightSpaceMatrix [6]mgl32.Mat4

	savedLocations map[uint32][]int32

	OrthoLimits []float32
}

func (light *Light) UpdateAttenuationCoefficients() {
	if light.Type == 0 {
		return
	}
	light.Constant = 1
	light.Linear = 2 / light.MaxDistance
	light.Quadratic = 1 / float32(math.Pow(float64(light.MaxDistance), 2))
}

func (light *Light) UpdateLightSpaceMatrix(camera *Camera) {
	switch light.Type {
	case 0:
		dir := light.Direction.Normalize()
		pos := camera.Position.Sub(dir.Mul(50))

		lightProj := mgl32.Ortho(light.OrthoLimits[0], light.OrthoLimits[1], light.OrthoLimits[2], light.OrthoLimits[3], .1, 100)
		lightView := mgl32.LookAtV(pos, camera.Position, mgl32.Vec3{0, 1, 0})

		light.LightSpaceMatrix[0] = lightProj.Mul4(lightView)
	case 1:
		pointLightDirections := [6]mgl32.Vec3{
			{1, 0, 0},
			{-1, 0, 0},
			{0, 1, 0},
			{0, -1, 0},
			{0, 0, 1},
			{0, 0, -1},
		}

		pointLightUps := [6]mgl32.Vec3{
			{0, -1, 0},
			{0, -1, 0},
			{0, 0, 1},
			{0, 0, -1},
			{0, -1, 0},
			{0, -1, 0},
		}

		lightProj := mgl32.Perspective(
			mgl32.DegToRad(90),
			1.0,
			1,
			light.MaxDistance,
		)

		for i, direction := range pointLightDirections {
			target := light.Position.Add(direction)

			up := pointLightUps[i]

			lightView := mgl32.LookAtV(light.Position, target, up)

			light.LightSpaceMatrix[i] = lightProj.Mul4(lightView)
		}
	case 2:
		dir := light.Direction

		target := light.Position.Add(dir)

		lightProj := mgl32.Perspective(
			light.OuterCutOut*2,
			1.0,
			1,
			light.MaxDistance,
		)

		up := mgl32.Vec3{0, 1, 0}
		if math.Abs(float64(dir.X())) < 0.0001 && math.Abs(float64(dir.Z())) < 0.0001 {
			up = mgl32.Vec3{0, 0, 1}
		}

		lightView := mgl32.LookAtV(light.Position, target, up)

		//fmt.Println(lightView)

		light.LightSpaceMatrix[0] = lightProj.Mul4(lightView)
	}
}

func (light *Light) SetUniform(shaderProgram ShaderProgram, uniform string) {
	locations, ok := light.savedLocations[shaderProgram.program]

	if !ok {
		locations = []int32{
			shaderProgram.GetUniformLocation(uniform + ".type"),

			shaderProgram.GetUniformLocation(uniform + ".constant"),
			shaderProgram.GetUniformLocation(uniform + ".linear"),
			shaderProgram.GetUniformLocation(uniform + ".quadratic"),

			shaderProgram.GetUniformLocation(uniform + ".position"),
			shaderProgram.GetUniformLocation(uniform + ".direction"),

			shaderProgram.GetUniformLocation(uniform + ".diffuse"),

			shaderProgram.GetUniformLocation(uniform + ".innerCutOut"),
			shaderProgram.GetUniformLocation(uniform + ".outerCutOut"),

			shaderProgram.GetUniformLocation(uniform + ".lightSpaceMatrix"),

			shaderProgram.GetUniformLocation(uniform + ".maxDistance"),
		}

		light.savedLocations[shaderProgram.program] = locations
	}

	gl.Uniform1i(locations[0], light.Type)

	gl.Uniform1f(locations[1], light.Constant)
	gl.Uniform1f(locations[2], light.Linear)
	gl.Uniform1f(locations[3], light.Quadratic)

	gl.Uniform3f(locations[4], light.Position[0], light.Position[1], light.Position[2])
	gl.Uniform3f(locations[5], light.Direction[0], light.Direction[1], light.Direction[2])

	gl.Uniform3f(locations[6], light.Diffuse[0], light.Diffuse[1], light.Diffuse[2])

	gl.Uniform1f(locations[7], float32(math.Cos(float64(light.InnerCutOut))))
	gl.Uniform1f(locations[8], float32(math.Cos(float64(light.OuterCutOut))))

	gl.UniformMatrix4fv(locations[9], 6, false, &light.LightSpaceMatrix[0][0])

	gl.Uniform1f(locations[10], light.MaxDistance)
}

func makeLightContainer(lightSources ...*Light) *LightContainer {
	return &LightContainer{
		lightSources,
	}
}

type LightContainer struct {
	LightSources []*Light
}

func (lightCont *LightContainer) GetType(ltype int32) []*Light {
	lights := []*Light{}

	for _, lightSource := range lightCont.LightSources {
		if lightSource.Type != ltype {
			continue
		}
		lights = append(lights, lightSource)
	}

	return lights
}

func (lightCont *LightContainer) GetSortedTypes(ltypesOrder []int32) []*Light {
	lights := make([]*Light, len(lightCont.LightSources))

	i := 0
	for _, ltype := range ltypesOrder {
		for _, lightSource := range lightCont.LightSources {
			if lightSource.Type != ltype {
				continue
			}
			lights[i] = lightSource
			i++
		}
	}

	return lights
}
