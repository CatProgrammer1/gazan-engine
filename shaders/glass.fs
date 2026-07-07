#version 430 core
out vec4 FragColor;

struct Material {
    sampler2D diffuse;
    sampler2D specular;
    float opacity;      
    float reflectance;  
    float shininess;    
}; 

uniform Material material;
uniform vec3 viewPos;

in vec3 FragPos;
in vec2 TexCoords;
in vec3 normal;

struct Light {
    int type; // 0 - Directional, 1 - Point
    vec3 position;
    vec3 direction;
    vec3 ambient;
    vec3 diffuse;
    vec3 specular;
    float constant;
    float linear;
    float quadratic;
};

#define MAX_LIGHTS 50
uniform int lightSourcesCount;
uniform Light lightSources[MAX_LIGHTS];  

float calcFresnel() {
    vec3 viewDir = normalize(viewPos - FragPos);
    vec3 norm = normalize(normal);

    float cosTheta = clamp(dot(norm, viewDir), 0.0, 1.0);
    
    float R0 = material.reflectance;
    
    return R0 + (1.0 - R0) * pow(1.0 - cosTheta, 5.0);
}

vec3 calculateLight(Light light, float fresnel) {
    return mix(vec3(texture(material.diffuse, TexCoords)), light.diffuse, fresnel);
}

void main() {
    float fresnel = calcFresnel();
    vec3 result;

    for (int i = 0; i < lightSourcesCount; i++) {
        if (i >= MAX_LIGHTS) break;
        result += calculateLight(lightSources[i], fresnel);
    }

    float finalAlpha = material.opacity + (1.0 - material.opacity) * fresnel;

    FragColor = vec4(result, finalAlpha);
}