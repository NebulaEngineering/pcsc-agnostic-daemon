# pcscrest

Conjunto de herraminetas PCSC expuestas a través de una API REST

## PRE-REQUISITOS

- El binario puede ser compilado en Linux(4 or later), Windows(10 o or later), o MAC(OS).
- Soporte PCSC en el Sistema Operativo (ejemplo: Linuc => pcsc-lite).
- Drivers PCSC de la lectora de tarjetas inteligentes.

## Compilación

1. Instalar GOLANG [https://go.dev/doc/install](https://go.dev/doc/install)

2. Descargar el proyecto

 `git clone https://gitlab.com/nebulaeng/fleet/pcscrest.git`

3. Moverse al directorio del binario que será creado

 `cd pcscrest/cmd/server`

4. Crear el binario

 `go build -o pcscrest .`

5. Copiar el binario en el directorio final desde el que será ejecutado. Ejemplo:

 `cp pcscrest ~/bin/` 

