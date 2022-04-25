package main

import "flag"

var filenameKey = "/key.pem"
var filenameCert = "/cert.pem"

func init() {
	flag.StringVar(&certpath, "certpath", "$HOME", "path to certificate file, if this option wasn't defined the application will create a new temporal certificatee")
}
