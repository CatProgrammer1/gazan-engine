package main

import (
	"math"

	"github.com/go-gl/glfw/v3.4/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

func newCamera(window *glfw.Window, position, front, up mgl32.Vec3, projection mgl32.Mat4) *Camera {
	camera := &Camera{
		Position: position,
		Front:    front.Normalize(),
		Up:       up,

		projection: projection,
	}

	camera.Update()

	var pitch, yaw float32

	firstMouse := true

	sensitivity := .1

	lastX, lastY := 0.0, 0.0

	window.SetCursorPosCallback(func(w *glfw.Window, xpos, ypos float64) {
		if firstMouse {
			firstMouse = false
			lastX = xpos
			lastY = ypos
		}

		xoffset := xpos - lastX
		yoffset := lastY - ypos
		lastX = xpos
		lastY = ypos

		xoffset *= sensitivity
		yoffset *= sensitivity

		yaw += float32(xoffset)
		pitch += float32(yoffset)

		if pitch > 89 {
			pitch = 89
		} else if pitch < -89 {
			pitch = -89
		}

		camera.Front = mgl32.Vec3{
			float32(
				math.Cos(float64(mgl32.DegToRad(yaw))) * math.Cos(float64(mgl32.DegToRad(pitch))),
			),
			float32(
				math.Sin(float64(mgl32.DegToRad(pitch))),
			),
			float32(
				math.Sin(float64(mgl32.DegToRad(yaw))) * math.Cos(float64(mgl32.DegToRad(pitch))),
			),
		}.Normalize()
	})

	return camera
}

type Camera struct {
	Position, Front, CameraRight, CameraUp mgl32.Vec3
	Up                                     mgl32.Vec3

	Type int //0 - basic

	projection,
	view mgl32.Mat4
}

func (camera *Camera) Update() {
	camera.CameraRight = camera.Up.Cross(camera.Front).Normalize()
	camera.CameraUp = camera.Front.Cross(camera.CameraRight)

	camera.view = mgl32.LookAtV(camera.Position, camera.Position.Add(camera.Front), camera.Up)
}

func (camera *Camera) SetPositionF(x, y, z float32) {
	camera.Position = mgl32.Vec3{x, y, z}
}

func (camera *Camera) SetPosition(vec mgl32.Vec3) {
	camera.Position = vec
}

func (camera *Camera) SetFrontF(x, y, z float32) {
	camera.Front = mgl32.Vec3{x, y, z}.Normalize()
}

func (camera *Camera) SetFront(vec mgl32.Vec3) {
	camera.Front = vec.Normalize()
}

func (camera *Camera) View() mgl32.Mat4 {
	return camera.view
}

func (camera *Camera) ViewPtr() *float32 {
	return &camera.view[0]
}
