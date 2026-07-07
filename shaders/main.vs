#version 430 core
layout(location = 0) in vec3 aPosition;
layout(location = 1) in vec3 aNormal;
layout(location = 2) in vec2 aTexCoords;
layout(location = 3) in ivec4 aBoneIDs;
layout(location = 4) in vec4 aWeights;
layout(location = 5) in vec4 aTangent;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;
uniform mat4 gBones[100];

out mat3 TBN;
out vec3 FragPos;
out vec2 TexCoords;
out vec3 normal;

void main() {
    vec4 posL = (gBones[aBoneIDs.x] * vec4(aPosition, 1.0)) * aWeights.x +
        (gBones[aBoneIDs.y] * vec4(aPosition, 1.0)) * aWeights.y +
        (gBones[aBoneIDs.z] * vec4(aPosition, 1.0)) * aWeights.z +
        (gBones[aBoneIDs.w] * vec4(aPosition, 1.0)) * aWeights.w;

    mat3 skinMatrix = mat3(gBones[aBoneIDs.x]) * aWeights.x +
        mat3(gBones[aBoneIDs.y]) * aWeights.y +
        mat3(gBones[aBoneIDs.z]) * aWeights.z +
        mat3(gBones[aBoneIDs.w]) * aWeights.w;

    mat3 normalMatrix = transpose(inverse(mat3(model)));

    vec3 N = normalize(normalMatrix * skinMatrix * aNormal);
    vec3 T = normalize(normalMatrix * skinMatrix * aTangent.xyz);

    T = normalize(T - dot(T, N) * N);

    vec3 B = cross(N, T) * aTangent.w;

    TBN = mat3(T, B, N);
    normal = N;

    TexCoords = aTexCoords;

    vec4 worldPos = model * posL;
    FragPos = worldPos.xyz;
    gl_Position = projection * view * worldPos;
}
