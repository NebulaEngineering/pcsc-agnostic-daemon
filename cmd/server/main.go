package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"gitlab.com/nebulaeng/fleet/pcscrest/internal/handler"
)

var certpath string
var keypath string
var create bool

func init() {
	// flag.StringVar(&certpath, "certpath", "$HOME", "path to certificate file, if this option wasn't defined the application will create a new temporal certificatee")
	flag.StringVar(&keypath, "keypath", "", "path to key file, if this option and \"certpath\" option weren't defined the application will create a new temporal certificate")
	flag.BoolVar(&create, "c", true, "Create files if they don't exist?")
}

func main() {

	flag.Parse()

	router := mux.NewRouter().StrictSlash(true)

	corsWrapper := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE"},
		AllowedHeaders: []string{"Content-Type", "Origin", "Accept", "*"},
	})

	router.
		Methods("GET").
		Path("/pcsc-daemon/service").
		Name("getState").
		HandlerFunc(handler.GetState)

	router.
		Methods("GET").
		Path("/pcsc-daemon/readers").
		Name("listCardReaderDevices").
		HandlerFunc(handler.ListAllReaders)

	router.
		Methods("GET").
		Path("/pcsc-daemon/readers/{id}").
		Name("getReaderInformation").
		HandlerFunc(handler.ReaderInformation)

	router.
		Methods("GET").
		Path("/pcsc-daemon/readers/{readerIdHex}/smartcard").
		Name("getReaderCards").
		HandlerFunc(handler.CardInReader)

	router.
		Methods("POST").
		Path("/pcsc-daemon/readers/{readerIdHex}/smartcard/sendApdus").
		Name("sendAPDUs").
		HandlerFunc(handler.SendAPUs)

	router.
		Methods("POST").
		Path("/pcsc-daemon/readers/{readerIdHex}/smartcard/sendApdu").
		Name("sendAPDU").
		HandlerFunc(handler.SendAPU)

	server := &http.Server{
		Addr:    ":1215",
		Handler: corsWrapper.Handler(router),
	}

	cert, key, err := verifyAndCreateFiles(certpath, keypath, create)
	if err != nil {
		log.Fatalln(err)
	}

	log.Fatal(server.ListenAndServeTLS(cert, key))
}
