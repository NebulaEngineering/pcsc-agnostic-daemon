package app

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/looplab/fsm"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/internal/pcsc/card"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/internal/pcsc/context"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/internal/pcsc/reader"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/utils"
)

type app struct {
	ctx           *context.Context
	cardsReader   map[string]*card.Card
	cardsSession  map[string]*card.Card
	sessionReader map[string]string
	frun          bool
	fmachine      *fsm.FSM
	mux           sync.Mutex
}

type App interface {
	ListReaders() ([]string, error)
	ReaderInformation(key string) (string, error)
	ConnectCardInReader(nameReader string) (*card.Card, error)
	VerifyCardInReader(nameReader string) (*card.Card, error)
	SendAPUs(nameReader, sessionId string, closeSession bool, data ...[]byte) (<-chan []byte, error)
}

var instance *app
var once sync.Once

//getInstance create App
func Instance() App {

	once.Do(func() {
		instance = &app{
			mux:           sync.Mutex{},
			cardsSession:  make(map[string]*card.Card),
			cardsReader:   make(map[string]*card.Card),
			sessionReader: make(map[string]string),
		}

		instance.fmachine = NewFSM()
		instance.runFSM()
		// var err error
		// instance.ctx, err = context.New()
		// if err != nil {
		// 	log.Println(err)
		// }
		time.Sleep(1 * time.Second)
	})
	return instance
}

func (app *app) ListReaders() ([]string, error) {
	app.mux.Lock()
	defer app.mux.Unlock()
	if app.ctx == nil {
		return nil, errors.New("context with reader is not valid")
	}
	rds, err := app.ctx.ListReaders()
	if err != nil {
		return nil, err
	}
	return rds, nil
}

func (app *app) ReaderInformation(key string) (string, error) {
	app.mux.Lock()
	defer app.mux.Unlock()
	if app.ctx == nil {
		return "", fmt.Errorf("smardcard context is nil")
	}
	if app.ctx == nil {
		return "", errors.New("context with reader is not valid")
	}
	return app.ctx.ReaderInformation(key)
}

func (app *app) ConnectCardInReader(nameReader string) (*card.Card, error) {
	app.mux.Lock()
	defer app.mux.Unlock()
	if app.ctx == nil {
		return nil, fmt.Errorf("smardcard context is nil")
	}

	r, err := reader.ConnectReader(app.ctx, nameReader)
	if err != nil {
		// fmt.Println("WWWWWW")
		return nil, err
	}

	cardx, err := card.ConnectCard(r)
	if err != nil {
		// fmt.Printf("ZZZZZZZ: %q", r.Name())
		if c, ok := app.cardsReader[r.Name()]; ok {
			c.Disconnect()
		}
		return nil, err
	}
	app.cardsReader[r.Name()] = cardx
	return cardx, nil
}

func (app *app) VerifyCardInReader(nameReader string) (*card.Card, error) {
	if v, ok := app.cardsReader[nameReader]; ok {
		// fmt.Println("XXXXXX")
		if _, err := v.Status(); err != nil {
			v.Disconnect()
			delete(app.cardsReader, nameReader)
			return nil, fmt.Errorf("error VerifyCardInReader: %w", err)
		}

		return v, nil
	}

	r, err := reader.ConnectReader(app.ctx, nameReader)
	if err != nil {
		// fmt.Println("YYYYYY")
		return nil, err
	}

	return app.ConnectCardInReader(r.Name())
}

func (app *app) SendAPUs(nameReader, sessionId string, closeSession bool, data ...[]byte) (<-chan []byte, error) {

	// fmt.Printf("data: %X\n", data)
	var cardx *card.Card
	var err error

	if c, ok := app.cardsSession[sessionId]; ok {
		cardx = c
	}
	if cardx == nil || func() bool {
		if _, err := cardx.Status(); err != nil {
			fmt.Printf("error status: %s\n", err)
			cardx.Disconnect()
			return true
		}
		return false
	}() {
		if v, ok := app.cardsReader[nameReader]; ok {
			v.Disconnect()
			delete(app.cardsReader, nameReader)
		}
		cardx, err = app.ConnectCardInReader(nameReader)
		if err != nil {
			fmt.Printf("erro: %s\n", err)
			return nil, err
		}
		// app.cardsReader[nameReader] = cardx
		app.cardsSession[sessionId] = cardx
		app.sessionReader[nameReader] = sessionId
	}
	// fmt.Printf("data: %X\n", data)

	ch := make(chan []byte)
	go func(cardz *card.Card, closeSs bool) {
		app.mux.Lock()
		var errF error
		defer func() {
			close(ch)
			if closeSs || errF != nil {
				if errF != nil {
					fmt.Println(err)
				}
				if cardz != nil {
					cardz.Disconnect()
				}
				delete(app.cardsSession, sessionId)
				delete(app.cardsReader, nameReader)
			}
			app.mux.Unlock()
		}()
		if app.ctx == nil {
			errF = fmt.Errorf("smardcard context is nil")
			return
		}
		if err := func() (errx error) {
			for _, d := range data {
				// fmt.Printf("data 1: %X\n", d)
				response, err := cardz.SendAPDU(d)
				if err != nil {
					return fmt.Errorf("error sendApud = %w", err)
				}
				select {
				case ch <- response:
					if len(d) > 0 && (d[0]&0x03 == 0x80 || d[0] == 0xFF) {
						// fmt.Printf("evaluation: [ %X ], [ %X ]\n", d, response)
						if !utils.VerifyResponseISO7816(response) {
							return fmt.Errorf("bad response: [% X]", response)
						}
					} else if !utils.VerifyResponse(response) {
						return fmt.Errorf("bad response: [% X]", response)
					}
				case <-time.After(3 * time.Second):
					return fmt.Errorf("sendApdu timeout")
				}
			}
			return nil
		}(); err != nil {
			errF = err
		}
	}(cardx, closeSession)
	return ch, nil
}
