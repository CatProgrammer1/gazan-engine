#version 430 core

uniform sampler2D frame_image;
in vec2 v_texCoord;

out vec4 finalColor;

void main() {
    vec2 pixelatedUV = floor(v_texCoord / .005f) * .005f;
    
    vec4 color = texture(frame_image, pixelatedUV);

    finalColor = color;
}
