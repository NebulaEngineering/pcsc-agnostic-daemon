package app

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"sync"
	"time"

	pcontext "context"

	"github.com/looplab/fsm"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/internal/pcsc/card"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/internal/pcsc/context"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/internal/pcsc/reader"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/utils"
)

var disableSession bool

func init() {
	flag.BoolVar(&disableSession, "disable-sessions", false, "disable session for card")
	// flag.Parse()
}

type app struct {
	contxt      pcontext.Context
	ctx         *context.Context
	cardsReader map[string]*card.Card
	// cardsSession  map[string]*card.Card
	// sessionReader map[string]string
	frun     bool
	fmachine *fsm.FSM
	mux      sync.Mutex
}

type App interface {
	ListReaders() ([]string, error)
	ReaderInformation(key string) (string, error)
	ConnectCardInReader(nameReader string) (*card.Card, error)
	VerifyCardInReader(nameReader string) (*card.Card, error)
	SendAPUs(nameReader, sessionId string, closeSession, debug bool, data ...[]byte) (<-chan []byte, error)
}

var instance *app
var once sync.Once
var lock sync.Mutex

// getInstance create App
func InitInstance(ctx pcontext.Context) App {

	lock.Lock()
	defer lock.Unlock()
	if instance == nil {
		instance = &app{
			contxt: ctx,
			mux:    sync.Mutex{},
			// cardsSession:  make(map[string]*card.Card),
			cardsReader: make(map[string]*card.Card),
			// sessionReader: make(map[string]string),
		}
	}

	instance.fmachine = NewFSM()
	instance.runFSM()
	// var err error
	// instance.ctx, err = context.New()
	// if err != nil {
	// 	log.Println(err)
	// }
	time.Sleep(1 * time.Second)

	// })
	return instance
}

func Instance() App {
	once.Do(func() {
		lock.Lock()
		defer lock.Unlock()
		if instance == nil {
			instance = &app{
				contxt: pcontext.Background(),
				mux:    sync.Mutex{},
				// cardsSession:  make(map[string]*card.Card),
				cardsReader: make(map[string]*card.Card),
				// sessionReader: make(map[string]string),
			}

			instance.fmachine = NewFSM()
			instance.runFSM()
			// var err error
			// instance.ctx, err = context.New()
			// if err != nil {
			// 	log.Println(err)
			// }
			time.Sleep(1 * time.Second)
		}
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
	if ok, err := app.ctx.IsValid(); err != nil || !ok {
		return nil, fmt.Errorf("context is not valid, err: %w", err)
	}

	r, err := reader.ConnectReader(app.ctx, nameReader)
	if err != nil {
		// fmt.Println("WWWWWW")
		return nil, fmt.Errorf("error ConnectReader: %w", err)
	}

	cardx, err := card.ConnectCard(r)
	if err != nil {
		// fmt.Printf("ZZZZZZZ: %q", r.Name())
		if c, ok := app.cardsReader[r.Name()]; ok {
			c.Disconnect()
		}
		return nil, fmt.Errorf("error ConnectCard: %w", err)
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
	if ok, err := app.ctx.IsValid(); err != nil || !ok {
		return nil, fmt.Errorf("context is not valid, err: %w", err)
	}

	r, err := reader.ConnectReader(app.ctx, nameReader)
	if err != nil {
		// fmt.Println("YYYYYY")
		return nil, fmt.Errorf("error ConnectReader: %w", err)
	}

	return app.ConnectCardInReader(r.Name())
}

func (app *app) SendAPUs(nameReader, sessionId string, closeSession, debug bool, data ...[]byte) (<-chan []byte, error) {

	// fmt.Printf("data: %X\n", data)
	var cardx *card.Card
	// var err error

	if err := func() error {
		if c, ok := app.cardsReader[nameReader]; !ok {
			return fmt.Errorf("card in reader not found (%s)", nameReader)
		} else {
			if _, err := c.Status(); err != nil {
				c.Disconnect()
				// delete(app.cardsReader, nameReader)
				return fmt.Errorf("error status: %w", err)
			} else {
				if !disableSession {
					if len(c.GetSessionID()) <= 0 {
						c.SetSessionID(sessionId)
					} else if !strings.EqualFold(sessionId, c.GetSessionID()) {
						c.Disconnect()
						return fmt.Errorf("session id not match (%s)", sessionId)
					}
					cardx = c
				} else {
					cardx = c
				}
				return nil
			}
		}
		return fmt.Errorf("error card not found")
	}(); err != nil {
		fmt.Println(err)
		card, err := app.ConnectCardInReader(nameReader)
		if err != nil {
			fmt.Printf("erro: %s\n", err)
			return nil, err
		}

		card.SetSessionID(sessionId)
		cardx = card

		app.cardsReader[nameReader] = cardx
	}

	// // app.cardsReader[nameReader] = cardx
	// app.cardsSession[sessionId] = cardx
	// app.sessionReader[nameReader] = sessionId

	// fmt.Printf("data: %X\n", data)

	ch := make(chan []byte)
	go func(cardz *card.Card, closeSs bool) {
		defer close(ch)
		app.mux.Lock()
		defer app.mux.Unlock()
		var errF error
		defer func() {
			if closeSs || errF != nil {
				if errF != nil {
					fmt.Println(errF)
				}
				if cardz != nil {
					cardz.Disconnect()
				}
				// delete(app.cardsSession, sessionId)
				delete(app.cardsReader, nameReader)
			}
		}()
		if app.ctx == nil {
			errF = fmt.Errorf("smardcard context is nil")
			return
		}
		if err := func() (errx error) {
			for _, d := range data {
				if debug {
					fmt.Printf("request APDU: [% X]\n", d)
				}
				response, err := cardz.SendAPDU(d)
				if err != nil {
					return fmt.Errorf("error sendApud = %w", err)
				}
				if debug {
					fmt.Printf("response APDU: [% X]\n", response)
				}
				select {
				case ch <- response:
					if len(d) > 0 && (d[0]&0xFC == 0x80 || d[0] == 0xFF) {
						// fmt.Printf("evaluation: [ %X ], [ %X ]\n", d, response)
						if !utils.VerifyResponseISO7816(response) {
							return fmt.Errorf("bad response: [% X]", response)
						}
					} else if !utils.VerifyResponse(response) {
						return fmt.Errorf("bad response: [% X]", response)
					}
				case <-time.After(1 * time.Second):
					return fmt.Errorf("sendApdu timeout")
				}
			}
			return nil
		}(); err != nil {
			fmt.Println(err)
			errF = err
		}
	}(cardx, closeSession)
	return ch, nil
}
