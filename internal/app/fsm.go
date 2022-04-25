package app

import (
	"fmt"
	"log"
	"time"

	"github.com/ebfe/scard"
	"github.com/looplab/fsm"
	"gitlab.com/nebulaeng/fleet/pcscrest/internal/pcsc/card"
	"gitlab.com/nebulaeng/fleet/pcscrest/internal/pcsc/context"
)

const (
	sOpen      = "sOpen"
	sList      = "sList"
	sClose     = "sClose"
	sRestart   = "sRestart"
	sWaitEvent = "sWait"
	sFatal     = "sFatal"
)

const (
	eOpenCmd = "eOpenCmd"
	eOpened  = "eOpened"
	eListed  = "eListed"
	eClosed  = "eClosed"
	eError   = "eError"
)

// func beforeEvent(event string) string {
// 	return fmt.Sprintf("before_%s", event)
// }

func enterState(state string) string {
	return fmt.Sprintf("enter_%s", state)
}

// func leaveState(state string) string {
// 	return fmt.Sprintf("leave_%s", state)
// }

func NewFSM() *fsm.FSM {

	calls := fsm.Callbacks{
		"enter_state": func(e *fsm.Event) {
			log.Printf("FSM SAM state Src: %v, state Dst: %v", e.Src, e.Dst)
		},
		"leave_state": func(e *fsm.Event) {
			if e.Err != nil {
				e.Cancel(e.Err)
			}
		},
		"before_event": func(e *fsm.Event) {
			if e.Err != nil {
				e.Cancel(e.Err)
			}
		},
		enterState(sOpen): func(e *fsm.Event) {
		},
		enterState(sClose): func(e *fsm.Event) {
		},
		enterState(sRestart): func(e *fsm.Event) {
		},
		enterState(sFatal): func(e *fsm.Event) {
		},
	}

	f := fsm.NewFSM(
		sClose,
		fsm.Events{
			{Name: eOpenCmd, Src: []string{sClose, sRestart}, Dst: sOpen},
			{Name: eOpened, Src: []string{sOpen}, Dst: sList},
			{Name: eListed, Src: []string{sList}, Dst: sWaitEvent},
			{Name: eClosed, Src: []string{sOpen, sWaitEvent}, Dst: sClose},
			{Name: eError, Src: []string{sWaitEvent, sList, sOpen}, Dst: sRestart},
			{Name: eError, Src: []string{sClose, sRestart}, Dst: sFatal},
		},
		calls,
	)
	return f
}

func (app *app) runFSM() {
	if app.frun {
		return
	}
	app.frun = true
	var readers []scard.ReaderState
	lastVerify := time.Now().Add(-30 * time.Second)
	go func() {
		defer func() {
			app.frun = false
		}()
		for {
			switch app.fmachine.Current() {
			case sOpen:
				func() {
					app.mux.Lock()
					defer app.mux.Unlock()
					ctx, err := context.New()
					if err != nil {
						log.Println(err)
						return
					}
					app.ctx = ctx
					// rds, err := ctx.ListReaders()
					// if err != nil {
					// 	log.Printf("readers error: %s", err)
					// 	return
					// }
					// readers = make([]scard.ReaderState, 0)
					// for _, r := range rds {
					// 	readers = append(readers, scard.ReaderState{
					// 		Reader:       r,
					// 		UserData:     nil,
					// 		CurrentState: scard.StateEmpty,
					// 		EventState:   scard.StateEmpty,
					// 		Atr:          nil,
					// 	})
					// }

					// if err := app.ctx.GetStatusChange(readers, 1*time.Second); err != nil {
					// 	log.Println(err)
					// 	return
					// }
					// log.Printf("readers state: %+v", readers)
					app.fmachine.Event(eOpened)
				}()
			case sList:
				func() {
					app.mux.Lock()
					defer app.mux.Unlock()
					if app.ctx == nil {
						log.Println("app context is nil")
						return
					}
					if time.Since(lastVerify) < 3*time.Second {
						return
					}
					lastVerify = time.Now()
					rds, err := app.ctx.ListReaders()
					if err != nil {
						log.Printf("readers error: %s", err)
						return
					}
					if len(rds) <= 0 {
						log.Printf("readers not found: %v", rds)
						return
					}
					readers = make([]scard.ReaderState, 0)
					for _, r := range rds {
						readers = append(readers, scard.ReaderState{
							Reader:       r,
							UserData:     nil,
							CurrentState: scard.StateEmpty,
							EventState:   scard.StateEmpty,
							Atr:          nil,
						})
					}

					if err := app.ctx.GetStatusChange(readers, 1*time.Second); err != nil {
						log.Println(err)
						return
					}
					log.Printf("readers state: %+v", readers)
					app.fmachine.Event(eListed)
				}()
			case sWaitEvent:
				func() {
					app.mux.Lock()
					defer app.mux.Unlock()
					if time.Since(lastVerify) > 3*time.Second {
						lastVerify = time.Now()
						if ok, err := app.ctx.IsValid(); err != nil || !ok {
							log.Printf("error context: %s, success: %v", err, ok)
							app.fmachine.Event(eClosed)
							return
						}
						if err := app.ctx.GetStatusChange(readers, 10*time.Millisecond); err != nil {
							log.Printf("status error: %s", err)
							app.fmachine.Event(eClosed)
							return
						}
						for _, r := range readers {
							switch r.EventState & 0xFF {
							case scard.StateEmpty,
								(scard.StateEmpty | scard.StateChanged),
								(scard.StateEmpty | scard.StateExclusive),
								(scard.StateEmpty | scard.StateChanged | scard.StateExclusive):
								// log.Printf("state: %X", r.EventState&0xFF)
								// log.Printf("reader: %+v", app.cardsReader)
								if v, ok := app.cardsReader[r.Reader]; ok {
									// if (r.EventState & scard.StateExclusive) != 0x00 {
									log.Printf("state: %X", r.EventState&0xFF)
									// }
									v.Disconnect()
									delete(app.cardsReader, r.Reader)
								}
								if v, ok := app.sessionReader[r.Reader]; ok {
									delete(app.sessionReader, r.Reader)
									delete(app.cardsSession, v)
								}
							}
						}
						// log.Printf("readers state: %+v", readers)
					}
				}()
			case sClose:
				func() {
					app.mux.Lock()
					defer app.mux.Unlock()
					app.cardsSession = make(map[string]*card.Card)
					app.cardsReader = make(map[string]*card.Card)
					app.sessionReader = make(map[string]string)
					if app.ctx != nil {
						if ok, err := app.ctx.IsValid(); err != nil || !ok {
							log.Printf("error context: %s, success: %v", err, ok)
						} else {
							app.ctx.Release()
						}
						app.ctx = nil
					}
				}()
				app.fmachine.Event(eOpenCmd)
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()
	time.Sleep(1 * time.Second)
}
