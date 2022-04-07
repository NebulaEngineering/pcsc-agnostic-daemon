package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"gitlab.com/nebulaeng/fleet/pcscrest/internal/handler"
)

func main() {

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
		Name("listCardReaderDevices").
		HandlerFunc(handler.SendAPUs)

	server := &http.Server{
		Addr:    ":1215",
		Handler: corsWrapper.Handler(router),
	}

	cert, key, err := newCert()
	if err != nil {
		log.Fatalf("create cert file error: %s", err)
	}

	buf := make([]byte, 1024)

	fcert, err := os.CreateTemp("", "certfile")
	if err != nil {
		log.Fatalf("create cert file error: %s", err)
	}

	for {
		if n, err := cert.Read(buf); err == nil {
			if _, err = fcert.Write(buf[:n]); err != nil {
				log.Fatalln(err)
			}
			continue
		}
		break
	}

	fkey, err := os.CreateTemp("", "keyfile")
	if err != nil {
		log.Fatalf("create key file error: %s", err)
	}

	for {
		if n, err := key.Read(buf); err == nil {
			if _, err = fkey.Write(buf[:n]); err != nil {
				log.Fatalln(err)
			}
			continue
		}
		break
	}

	log.Fatal(server.ListenAndServeTLS(fcert.Name(), fkey.Name())) //Key and cert are coming from Let's Encrypt
}
