package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"gitlab.com/nebulaeng/fleet/pcscrest/internal/handler"
)

var certpath string
var keypath string
var port int
var notcreate bool

func init() {
	flag.StringVar(&certpath, "certpath", "", "path to certificate file, if this option wasn't defined the application will create a new certificate in \"$HOME\"")
	flag.StringVar(&keypath, "keypath", "", "path to key file, if this option and \"certpath\" option weren't defined the application will create a new pair key in \"$HOME\"")
	flag.BoolVar(&notcreate, "f", false, "don't Create files if they don't exist?")
	flag.IntVar(&port, "port", 1025, "port in local socket to LISTEN (socket = localhost:port)")
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
		Addr:    fmt.Sprintf(":%d", port),
		Handler: corsWrapper.Handler(router),
	}

	if len(certpath) <= 0 {
		certpath = func() string {
			usr, err := user.Current()
			if err == nil {
				return usr.HomeDir
			}
			return os.Getenv("HOME")
		}()
	}
	cert, key, err := verifyAndCreateFiles(certpath, keypath, !notcreate)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("pcscrest starting ...")
	fmt.Println("pcscrest waiting for requests ...")
	log.Fatal(server.ListenAndServeTLS(cert, key))
}
