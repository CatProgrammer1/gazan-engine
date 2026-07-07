#version 430 core
out vec4 color;

#define PI 3.14159265359

struct Material {
    sampler2D diffuse;
    sampler2D normal;
    sampler2D metallicRoughness;

    float roughness;
    float metallic;
};

uniform Material material;

uniform samplerCube environment;

in vec3 FragPos;
in vec2 TexCoords;
in vec3 normal;
in mat3 TBN;

uniform vec3 viewPos;

const vec4 GLOBAL_AMBIENT = vec4(0.25, 0.25, 0.25, 1.0);

struct Light {
    int type; //0 - direction light, 1 - point light

    vec3 position;
    vec3 direction;

    vec3 diffuse;

    float constant;
    float linear;
    float quadratic;

    float innerCutOut;
    float outerCutOut;

    float maxDistance;

    mat4[6] lightSpaceMatrix;
};

uniform sampler2DArray shadowMapArray1;
uniform sampler2DArray shadowMapArray2;
uniform samplerCubeArray shadowMapArray3;

#define MAX_LIGHTS 10

uniform int lightSourcesCount;
uniform Light lightSources[MAX_LIGHTS];

float DistributionGGX(vec3 N, vec3 H, float roughness) {
    float a = roughness * roughness;
    float a2 = a * a; // По спецификации Disney используется четвертая степень для лучшего вида
    float NdotH = max(dot(N, H), 0.0);
    float NdotH2 = NdotH * NdotH;

    float nom = a2;
    float denom = (NdotH2 * (a2 - 1.0) + 1.0);
    denom = PI * denom * denom; // Делим на Пи

    return nom / denom;
}

float GeometrySchlickGGX(float NdotV, float roughness) {
    // Для прямых источников света (Analytical Lights) коэффициент k считается так:
    float r = (roughness + 1.0);
    float k = (r * r) / 8.0;

    float nom = NdotV;
    float denom = NdotV * (1.0 - k) + k;

    return nom / denom;
}

float GeometrySmith(vec3 N, vec3 V, vec3 L, float roughness) {
    float NdotV = max(dot(N, V), 0.0);
    float NdotL = max(dot(N, L), 0.0);
    float ggx2 = GeometrySchlickGGX(NdotV, roughness);
    float ggx1 = GeometrySchlickGGX(NdotL, roughness);

    return ggx1 * ggx2;
}

vec3 fresnelSchlick(float cosTheta, vec3 F0) {
    return F0 + (1.0 - F0) * pow(clamp(1.0 - cosTheta, 0.0, 1.0), 5.0);
}

vec3 getAmbientIBL(vec3 viewDir) {
    vec4 mrSample = texture(material.metallicRoughness, TexCoords);

    float roughness = clamp(mrSample.g * material.roughness, 0.05, 1.0);
    float metallic = mrSample.b * material.metallic;

    vec3 albedo = texture(material.diffuse, TexCoords).rgb;
    vec3 N = normalize(normal);
    vec3 V = normalize(viewDir);

    float NdotV = max(dot(N, V), 0.0);

    vec3 F0 = mix(vec3(0.04), albedo, metallic);
    vec3 R = reflect(-V, N);

    vec3 skyboxColor = pow(texture(environment, R).rgb, vec3(2.2));

    vec3 F_ambient = fresnelSchlick(NdotV, F0);
    vec3 kD_ambient = (vec3(1.0) - F_ambient) * (1.0 - metallic);

    float metalMask = mix(0.0, 1.0, metallic);

    vec3 envSpecular = skyboxColor * F_ambient * (1.0 - roughness) * metalMask;
    vec3 envDiffuse = GLOBAL_AMBIENT.rgb * albedo * kD_ambient;

    return envDiffuse + envSpecular;
}

//😭🙏
vec3 getDirectPBR(vec3 lightDir, vec3 lightColor, vec3 viewDir) {
    vec4 mrSample = texture(material.metallicRoughness, TexCoords);

    float roughness = clamp(mrSample.g * material.roughness, 0.05, 1.0);
    float metallic = mrSample.b * material.metallic;

    vec3 albedo = texture(material.diffuse, TexCoords).rgb;
    vec3 N = normalize(normal);

    vec3 V = normalize(viewDir);
    vec3 L = normalize(lightDir);
    vec3 H = normalize(V + L);

    float NdotV = max(dot(N, V), 0.0);
    float NdotL = max(dot(N, L), 0.0);

    vec3 F0 = mix(vec3(0.04), albedo, metallic);

    float D = DistributionGGX(N, H, roughness);
    float G = GeometrySmith(N, V, L, roughness);
    vec3 F = fresnelSchlick(max(dot(H, V), 0.0), F0);

    vec3 numerator = D * G * F;
    float denominator = 4.0 * NdotV * NdotL + 0.001;
    vec3 specularBRDF = numerator / denominator;

    vec3 kS = F;
    vec3 kD = (vec3(1.0) - kS) * (1.0 - metallic);

    vec3 diffuseBRDF = albedo / PI;

    return (kD * diffuseBRDF + specularBRDF) * lightColor * NdotL;
}

//Idk how this shit really works, but at least it works
float CalculateShadow(sampler2DArray shadowMapArray, float pcfRadius, vec4 fragPosLightSpace, vec3 norm, vec3 lightDir, int layer) {
    vec3 projCoords = fragPosLightSpace.xyz / fragPosLightSpace.w;
    projCoords = projCoords * 0.5 + 0.5;
    if(projCoords.z > 1.0)
        return 0.0;
    if(projCoords.x < 0.0 || projCoords.x > 1.0 ||
        projCoords.y < 0.0 || projCoords.y > 1.0)
        return 0.0;

    float currentDepth = projCoords.z;
    float cosTheta = max(dot(norm, lightDir), 0.0);
    float bias = max(0.002 * (1.0 - cosTheta), 0.00025);

    vec2 baseUV = projCoords.xy;

    vec2 texelSize = 1.0 / vec2(textureSize(shadowMapArray, 0).xy);

    float shadow = 0.0;
    int samples = 0;
    for(int x = -2; x <= 2; ++x) {
        for(int y = -2; y <= 2; ++y) {
            vec2 offset = vec2(x, y) * texelSize * pcfRadius;
            vec3 uvw = vec3(baseUV + offset, float(layer));
            float closestDepth = texture(shadowMapArray, uvw).r;
            shadow += (currentDepth - bias) > closestDepth ? 1.0 : 0.0;
            samples++;
        }
    }

    return shadow / float(samples);
}

float CalculateShadowCube(samplerCubeArray shadowMapArray, float pcfRadius, vec3 fragPos, vec3 lightPos, vec3 norm, float farPlane, int layer) {
    vec3 lightToFrag = fragPos - lightPos;
    float currentDepth = length(lightToFrag);

    if(currentDepth > farPlane)
        return 0.0;

    vec3 lightDir = normalize(-lightToFrag);
    float cosTheta = max(dot(norm, lightDir), 0.0);
    float bias = max(0.15 * (1.0 - cosTheta), 0.02); 

    float shadow = 0.0;
    int samples = 0;

    ivec3 texSize = textureSize(shadowMapArray, 0);
    float texelSize = 1.0 / float(texSize.x);
    float offsetStep = texelSize * pcfRadius;

    int max_coord = 1;

    for(int x = -max_coord; x <= max_coord; ++x) {
        for(int y = -max_coord; y <= max_coord; ++y) {
            for(int z = -max_coord; z <= max_coord; ++z) {
                vec3 offset = vec3(x, y, z) * offsetStep;
                vec3 sampleDir = lightToFrag + offset;

                vec4 sampleCoords = vec4(sampleDir, float(layer));
                float closestDepth = texture(shadowMapArray, sampleCoords).r;

                closestDepth *= farPlane;

                shadow += (currentDepth - bias) > closestDepth ? 1.0 : 0.0;
                samples++;
            }
        }
    }

    return shadow / float(samples);
}


vec3 calcDirLight(Light light, float shadow) {
    vec3 lightDir = normalize(-light.direction);
    vec3 viewDir = normalize(viewPos - FragPos);
    vec3 direct = getDirectPBR(lightDir, light.diffuse, viewDir);
    return (1.0 - shadow) * direct;
}

vec3 calcSpotLight(Light light, float shadow) {
    vec3 lightDir = normalize(light.position - FragPos);
    vec3 viewDir = normalize(viewPos - FragPos);

    float theta = dot(lightDir, normalize(-light.direction));
    float intensity = smoothstep(light.outerCutOut, light.innerCutOut, theta);
    intensity = pow(intensity, 2.0);

    float distance = length(light.position - FragPos);
    float attenuation = 1.0 / max(0.001, light.constant + light.linear * distance + light.quadratic * (distance * distance));

    vec3 direct = getDirectPBR(lightDir, light.diffuse, viewDir) * intensity * attenuation;
    return (1.0 - shadow) * direct;
}

vec3 calcPointLight(Light light, float shadow) {
    vec3 lightDir = normalize(light.position - FragPos);
    vec3 viewDir = normalize(viewPos - FragPos);

    float distance = length(light.position - FragPos);
    float attenuation = 1.0 / max(0.001, light.constant + light.linear * distance + light.quadratic * (distance * distance));

    vec3 direct = getDirectPBR(lightDir, light.diffuse, viewDir) * attenuation;
    return (1.0 - shadow) * direct;
}

void main() {
    vec3 norm = normalize(normal);
    vec3 viewDir = normalize(viewPos - FragPos);

    vec3 result = getAmbientIBL(viewDir);

    int dirI = 0;
    int spotI = 0;
    int pointI = 0;
    for (int i = 0; i < MAX_LIGHTS; i++) {
        if (i >= lightSourcesCount) break;
        Light light = lightSources[i];

        if (light.type == 0) {
            vec4 fragPosLightSpace = light.lightSpaceMatrix[0] * vec4(FragPos, 1.0);

            float shadow = CalculateShadow(shadowMapArray1, .5, fragPosLightSpace, norm, -light.direction, dirI);
            dirI++;
            result += calcDirLight(light, shadow);
        } else if (light.type == 1) {
            float shadow = CalculateShadowCube(shadowMapArray3, 10, FragPos, light.position, norm, light.maxDistance, pointI);

            result += calcPointLight(light, shadow);

            pointI++;
        } else if (light.type == 2) {
            vec4 fragPosLightSpace = light.lightSpaceMatrix[0] * vec4(FragPos, 1.0);

            float shadow = CalculateShadow(shadowMapArray2, .5, fragPosLightSpace, norm, -light.direction, spotI);
            spotI++;
            result += calcSpotLight(light, shadow);
        }
    }

    color = vec4(result, 1.0);
}