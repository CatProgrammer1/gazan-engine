#version 430 core
layout (location = 0) in vec3 aPosition;
layout (location = 3) in ivec4 aBoneIDs;
layout (location = 4) in vec4 aWeights;

uniform mat4 model;
uniform mat4 lightSpaceMatrix;
uniform mat4 gBones[100];

void main()
{
    mat4 BoneTransform = gBones[aBoneIDs.x] * aWeights.x;
    BoneTransform     += gBones[aBoneIDs.y] * aWeights.y;
    BoneTransform     += gBones[aBoneIDs.z] * aWeights.z;
    BoneTransform     += gBones[aBoneIDs.w] * aWeights.w;

    vec4 localPos = BoneTransform * vec4(aPosition, 1.0f);

    gl_Position = lightSpaceMatrix * model * localPos;
}
