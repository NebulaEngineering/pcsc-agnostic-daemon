package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/internal/app"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/internal/handler"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/utils"
	"github.com/rs/cors"
)

var certpath string
var keypath string
var port int
var notcreate bool
var showversion bool
var isdebug bool
var ssl bool

func init() {
	flag.StringVar(&certpath, "certpath", "", "[ssl enable required] path to certificate file, if this option wasn't defined the application will create a new certificate in \"$HOME\"")
	flag.StringVar(&keypath, "keypath", "", "[ssl enable required] path to key file, if this option and \"certpath\" option weren't defined the application will create a new pair key in \"$HOME\"")
	flag.BoolVar(&notcreate, "f", false, "don't Create files if they don't exist?")
	flag.BoolVar(&ssl, "ssl", false, "enable ssl local service?")
	flag.BoolVar(&showversion, "version", false, "show version")
	flag.BoolVar(&isdebug, "debug", false, "show APDUs in stdout")
	flag.IntVar(&port, "port", 1216, "port in local socket to LISTEN (socket = localhost:port)")
}

func main() {

	flag.Parse()
	if showversion {
		fmt.Printf("version: %s\n", handler.VERSION)
		os.Exit(2)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	utils.Debug = isdebug

	app.InitInstance(ctx)

	svc := &myService{
		ctx:        ctx,
		cancelFunc: cancel,
	}

	interactive := IsAnInteractiveSession()

	go func() {
		runService("pcsc-pos", interactive, svc)
	}()

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

	serverHttp := &http.Server{
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

	fmt.Println("pcsc-agnostic-daemon starting ...")
	fmt.Println("pcsc-agnostic-daemon waiting for requests ...")

	var serverSSL *http.Server

	if ssl {
		cert, key, err := verifyAndCreateFiles(certpath, keypath, !notcreate)
		if err != nil {
			log.Fatalln(err)
		}
		go func() {
			serverSSL = &http.Server{
				Addr:    fmt.Sprintf(":%d", port-1),
				Handler: corsWrapper.Handler(router),
			}

			if err := serverSSL.ListenAndServeTLS(cert, key); err != nil && err != http.ErrServerClosed {
				log.Fatalf("listenAndServerTLS: %s\n", err)
			}
		}()
	}

	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	go func() {
		fmt.Println("pcsc-agnostic-daemon net started ...")
		if err := serverHttp.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listenAndServer: %s\n", err)
		}
	}()

	for {
		select {
		case <-finish:
			ctxx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			if err := serverHttp.Shutdown(ctxx); err != nil {
				log.Printf("shutdown serverHttp: %s\n", err)
			}
			if ssl && serverSSL != nil {
				if err := serverSSL.Shutdown(ctxx); err != nil {
					log.Printf("shutdown serverSSL: %s\n", err)
				}
			}
			log.Println("Servidor cerrado correctamente")
			return
		}
	}

}
