# pcscrest

Conjunto de herraminetas PCSC expuestas a través de una API REST

## PRE-REQUISITOS

- Soporte PCSC en el Sistema Operativo (ejemplo: Linux => pcsc-lite).
- Drivers PCSC de la lectora de tarjetas inteligentes.
- El binario puede ser compilado en Linux(4 or later), Windows(10 o or later), o MAC(OS).


## Compilación

1. Instalar GOLANG [https://go.dev/doc/install](https://go.dev/doc/install)

2. Descargar el proyecto

```git clone https://gitlab.com/nebulaeng/fleet/pcscrest.git````

3. Moverse al directorio del binario que será creado

```cd pcscrest/cmd/server```

4. Crear el binario

```go build -o pcscrest .```

5. Copiar el binario en el directorio final desde el que será ejecutado. Ejemplo:

```cp pcscrest ~/bin/```

## Script de inicio [opcional]

A continuación se expone un ejemplo de la configuración de un Script de Inicio para el binario "pcscrest" en un sistema operativo Ubuntu a través de "systemd".

premisas del ejemplo:

- Existe un usuario en el sistema con el username "test"
- El directorio "home" del usaurio "test" en el sistema es "/home/test"
- El binario "pcscrest" existe en la ruta "/home/test/bin/pcscrest" y teien permisos de ejecucón para el usaurio "test".
- Hay un demonio de PCSC instalado ene l sistema (ejemplo: `sudo apt-get install pcscd`).
- Los drivers de la lectora de tarjetas sin contacto están instalados ene l sistema (ejemplo: `sudo apt-get install libacsccid1`). Es posible que sea necesario instalar los dirvers desde un "paquete" del fabricante de las lectoras. 

Script de inicio:

filename: /etc/systemd/system/pcscrest.service

```
# Simple service unit file to use for pcscrest
# startup configurations with systemd.
# By NebulaE
# Licensed under GPL V2
#

[Unit]
Description=PCSC local API REST

[Service]
Type=symple
Restart=always
RestartSec=3
User=test
ExecStart=/home/test/bin/pcscrest

[Install]
WantedBy=multi-user.target
```

Iniciar manualmente el Script en el sistema:

```systemd start pcscrest.service```

Detener manualmente el Script en el sistema:

```systemd stop pcscrest.service```

Habilita la ejecuación automática del Script cuando se inicie el sistema:

```systemd enable pcscrest.service```

Revisar el estado del Script:

```systemd status pcscrest.service```

```systemd enable pcscrest.service```


