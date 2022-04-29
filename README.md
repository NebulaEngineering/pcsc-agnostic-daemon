# pcscrest

Conjunto de herraminetas PCSC expuestas a través de una API REST

## Compilación

0. Instalar los drives PCSC de la lectora de tarjetas inteligentes.

1. Instalar GOLANG [https://go.dev/doc/install](https://go.dev/doc/install)

2. Descargar el proyecto

 `git clone https://gitlab.com/nebulaeng/fleet/pcscrest.git`

3. Moverse al directorio del binario que será creado

 `cd pcscrest/cmd/server`

4. Crear el binario

 `go build -o pcscrest .`

5. Copiar el binario en el directorio final desde el que será ejecutado. Ejemplo:

 `cp pcscrest ~/bin/` 

