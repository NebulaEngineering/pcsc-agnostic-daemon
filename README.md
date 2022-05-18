# PCSC AGNOSTIC DAEMON

Conjunto de herraminetas PCSC expuestas a través de una API REST

## PRE-REQUISITOS

- Soporte PCSC en el Sistema Operativo (ejemplo: Linux => pcsc-lite).
- Drivers PCSC de la lectora de tarjetas inteligentes.
- El binario puede ser compilado en Linux(4 or later), Windows(10 o or later), o MAC(OS).


## Compilación

1. Instalar GOLANG [https://go.dev/doc/install](https://go.dev/doc/install)

2. Descargar el proyecto

`git clone https://github.com/nebulaengineering/pcsc-agnostic-daemon.git`

3. Moverse al directorio del binario que será creado

`cd pcsc-agnostic-daemon/cmd/server`

4. Crear el binario

`go build -o pcsc-agnostic-daemon .`

5. (OPTIONAL) Copiar el binario en el directorio final desde el que será ejecutado. Ejemplo:

`cp pcsc-agnostic-daemon ~/bin/`

6. (OPCIONAL) se puede hacer una instalación "directa" desde "go" con la instrucción:

`go install https://github.com/nebulaengineering/pcsc-agnostic-daemon/cmd/server@latest``

El binario será instalado en el directorio "$GOPATH/bin" con el nombre original del paquete ("server" en este caso) "$GOPATH/go/bin/server. Se recomienda copiar el binario en la ruta de los binarios del usuario del sistema "~/bin" ("$HOME/bin").

`cp $GOPATH/bin/server ~/bin/pcsc-agnostic-daemon`

## Ejecución

A continuación se presentan las opciones de ejecución del binario.

- opciones:

```
pcsc-agnostic-daemon --help
Usage of pcsc-agnostic-daemon:
  -certpath string
    	path to certificate file, if this option wasn't defined the application will create a new certificate in "$HOME"
  -f	don't Create files if they don't exist?
  -keypath string
    	path to key file, if this option and "certpath" option weren't defined the application will create a new pair key in "$HOME"
  -port int
    	port in local socket to LISTEN (socket = localhost:port) (default 1215)
```



- Ejemplo de ejeución manual:

```
./pcsc-agnostic-daemon

pcsc-agnostic-daemon starting ...
pcsc-agnostic-daemon waiting for requests ...
```

La ejecución del binario sin opciones hará que éste busque los archivos del certificado y la llave TLS en las rutas "$HOME/cert.pem" y "$HOME/key.pem" respectivamente. Si estos archivos no existen el binario creará un par de llaves y un certificado autofirmado para el servico TLS dispuesto en el socket "localhost:port".


Ingrese a la siguiente URL con un browser para verificar la correcta ejecuación del binario:

[https://localhost:1215/pcsc-daemon/readers](https://localhost:1215/pcsc-daemon/readers)

Debería ver un listado de lectoras PCSC conectadas.

Si se hace uso del certificado creado automáticamente por el binario, es decir si no se usa un certifcado privado creado pr la organización, será necesario agregar el certificado creado (por defecto en la ruta "$HOME/cert.pem") al sistema de confianza del sistema operativo (probablemente instalando el certificado en el sistema) y habilitar la confianza en certificados digitales autofirmados para localhost.

Ejemplo de la habilitación de certificado TLS para localhost en chrome:

[chrome://flags](chrome://flags)

![flag_chrome](img/flag_chrome.png)


## Script de inicio [opcional]

A continuación se expone un ejemplo de la configuración de un Script de Inicio para el binario "pcsc-agnostic-daemon" en un sistema operativo Ubuntu a través de "systemd".

premisas del ejemplo:

- Existe un usuario en el sistema con el username "test"
- El directorio "home" del usaurio "test" en el sistema es "/home/test"
- El binario "pcsc-agnostic-daemon" existe en la ruta "/home/test/bin/pcsc-agnostic-daemon" y tiene permisos de ejecución para el usuario "test".
- Hay un demonio de PCSC instalado en el sistema (ejemplo: `sudo apt-get install pcscd`).
- Los drivers de la lectora de tarjetas sin contacto están instalados en el sistema (ejemplo: `sudo apt-get install libacsccid1`). Es posible que sea necesario instalar los dirvers desde un "paquete" del fabricante de las lectoras. 

Script de inicio:

filename: /etc/systemd/system/pcsc-agnostic-daemon.service

```
# Simple service unit file to use for pcsc-agnostic-daemon
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
ExecStart=/home/test/bin/pcsc-agnostic-daemon

[Install]
WantedBy=multi-user.target
```

Iniciar manualmente el Script en el sistema:

```systemctl start pcsc-agnostic-daemon.service```

Detener manualmente el Script en el sistema:

```systemctl stop pcsc-agnostic-daemon.service```

Habilita la ejecuación automática del Script cuando se inicie el sistema:

```systemctl enable pcsc-agnostic-daemon.service```

Revisar el estado del Script:

```systemctl status pcsc-agnostic-daemon.service```


