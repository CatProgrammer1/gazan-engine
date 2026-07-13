package main

import (
	"log"

	"github.com/go-gl/gl/v4.3-core/gl"
)

var (
	ppFBOs     = [2]uint32{}
	ppTextures = [2]uint32{}

	mainPPFBO, mainPPRBO, mainPPTexture uint32

	ppReady = false

	ppEnabled = false
)

func initPostProcessing(w, h int32) {
	if !ppReady {
		gl.GenFramebuffers(2, &ppFBOs[0])

		gl.GenTextures(2, &ppTextures[0])

		for i, tex := range ppTextures {
			gl.BindTexture(gl.TEXTURE_2D, tex)

			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

			gl.TexImage2D(
				gl.TEXTURE_2D,
				0,
				gl.RGBA,
				w,
				h,
				0,
				gl.RGBA,
				gl.UNSIGNED_BYTE,
				nil,
			)

			gl.BindFramebuffer(gl.FRAMEBUFFER, ppFBOs[i])

			gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, tex, 0)

			status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)
			if status != gl.FRAMEBUFFER_COMPLETE {
				panic(status)
			}
		}

		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		gl.BindTexture(gl.TEXTURE_2D, 0)

		gl.GenFramebuffers(1, &mainPPFBO)
		gl.BindFramebuffer(gl.FRAMEBUFFER, mainPPFBO)

		gl.GenTextures(1, &mainPPTexture)

		gl.BindTexture(gl.TEXTURE_2D, mainPPTexture)

		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

		gl.TexImage2D(
			gl.TEXTURE_2D,
			0,
			gl.RGBA,
			w,
			h,
			0,
			gl.RGBA,
			gl.UNSIGNED_BYTE,
			nil,
		)

		gl.FramebufferTexture(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, mainPPTexture, 0)

		gl.GenRenderbuffers(1, &mainPPRBO)
		gl.BindRenderbuffer(gl.RENDERBUFFER, mainPPRBO)

		gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH24_STENCIL8, w, h)

		gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_STENCIL_ATTACHMENT, gl.RENDERBUFFER, mainPPRBO)

		status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)
		if status != gl.FRAMEBUFFER_COMPLETE {
			log.Fatalln("Sigma lol framebuffer, shit", status)
		}

		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		gl.BindTexture(gl.TEXTURE_2D, 0)

		ppReady = true
	}
}

type PostProcess struct {
	ShaderProgram ShaderProgram
	TextureUnit   uint32
}

func (postProcess PostProcess) Bind(uniform string) {
	shaderProgram := postProcess.ShaderProgram

	gl.Uniform1i(shaderProgram.GetUniformLocation(uniform), int32(postProcess.TextureUnit-gl.TEXTURE0))
}

func newPostProcess(shaderProgram ShaderProgram, textureUnit uint32) *PostProcess {
	postProcess := &PostProcess{}

	postProcess.ShaderProgram = shaderProgram
	postProcess.TextureUnit = textureUnit

	return postProcess
}
