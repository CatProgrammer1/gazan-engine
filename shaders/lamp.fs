#version 430 core
out vec4 color;

uniform sampler2D LightOfTheWorld;

in vec2 TexCoords;

void main()
{
    color = vec4(1.0f)*texture(LightOfTheWorld, TexCoords);
}