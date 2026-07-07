#version 430 core
layout (triangles) in;
layout (triangle_strip, max_vertices=18) out;

uniform mat4 lightSpaceMatrix[6]; // Масив переїхав сюди з вершинного шейдера
uniform int cubeIndex; 

out vec4 FragPos; // Передаємо позицію у фрагментний шейдер для розрахунку відстані

void main()
{
    for(int face = 0; face < 6; ++face)
    {
        gl_Layer = (cubeIndex * 6) + face; // Вказуємо грань кубічної карти
        
        for(int i = 0; i < 3; ++i)
        {
            // gl_in[i].gl_Position — це світова позиція, яку ми порахували у VS
            FragPos = gl_in[i].gl_Position; 
            
            // Множимо світову позицію на матрицю конкретної грані світла
            gl_Position = lightSpaceMatrix[face] * FragPos;
            
            EmitVertex();
        }
        EndPrimitive();
    }
}
