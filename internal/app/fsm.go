package app

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"log"

	"github.com/ebfe/scard"
	"github.com/looplab/fsm"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/internal/pcsc/context"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/internal/pcsc/reader"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/utils"
)

const (
	sOpen    = "sOpen"
	sList    = "sList"
	sClose   = "sClose"
	sRestart = "sRestart"
	sWait    = "sWait"
	sFatal   = "sFatal"
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
			fmt.Printf("FSM SAM state Src: %v, state Dst: %v\n", e.Src, e.Dst)
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
		enterState(sWait): func(e *fsm.Event) {
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
			{Name: eOpenCmd, Src: []string{sClose, sRestart, sFatal}, Dst: sOpen},
			{Name: eOpened, Src: []string{sOpen}, Dst: sList},
			{Name: eListed, Src: []string{sList}, Dst: sWait},
			{Name: eClosed, Src: []string{sOpen, sWait}, Dst: sClose},
			{Name: eError, Src: []string{sWait, sList, sOpen}, Dst: sRestart},
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
	currentState := ""
	go func() {
		defer func() {
			app.frun = false
		}()
		tick := time.NewTicker(300 * time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case <-app.contxt.Done():
				log.Printf("context done (APP)")
				return
			case <-tick.C:
				if currentState != app.fmachine.Current() {
					currentState = app.fmachine.Current()
					fmt.Printf("currentState: %s\n", currentState)
				}
				switch app.fmachine.Current() {
				case sOpen:
					func() {
						app.mux.Lock()
						defer app.mux.Unlock()
						ctx, err := context.New()
						if err != nil {
							fmt.Println(err)
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
					if err := func() error {
						app.mux.Lock()
						defer app.mux.Unlock()
						if app.ctx == nil {
							return fmt.Errorf("app context is nil")
						}
						if time.Since(lastVerify) < 3*time.Second {

							return fmt.Errorf("time verify: %v", time.Since(lastVerify))
						}
						lastVerify = time.Now()
						rds, err := app.ctx.ListReaders()
						if err != nil {

							return fmt.Errorf("readers error: %s", err)
						}
						if len(rds) <= 0 {
							// fmt.Printf("readers not found: %v\n", rds)
							return fmt.Errorf("readers not found: %v", rds)
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

						for _, r := range rds {
							if strings.Contains(r, "ACS") {
								if err := reader.PrepareReader(app.ctx, r); err != nil {
									fmt.Println(err)
								}
								break
							}
						}

						if err := app.ctx.GetStatusChange(readers, 1*time.Second); err != nil {
							if !errors.Is(err, scard.ErrTimeout) {
								fmt.Println(err)
							}
							// return
						}
						fmt.Printf("readers state: %+v", readers)
						app.fmachine.Event(eListed)
						return nil
					}(); err != nil {
						fmt.Println(err)
						app.fmachine.Event(eError)
					}
				case sWait:
					func() {
						app.mux.Lock()
						defer app.mux.Unlock()
						if time.Since(lastVerify) > 3*time.Second {
							lastVerify = time.Now()
							if ok, err := app.ctx.IsValid(); err != nil || !ok {
								fmt.Printf("error context: %s, success: %v\n", err, ok)
								app.fmachine.Event(eClosed)
								return
							}
							if err := app.ctx.GetStatusChange(readers, 10*time.Millisecond); err != nil {
								if !errors.Is(err, scard.ErrTimeout) {
									fmt.Printf("status error: %s\n", err)
									app.fmachine.Event(eClosed)
									return
								}
							}
							for _, r := range readers {
								if utils.Debug {
									fmt.Printf("reader (%q) state: %02X\n", r.Reader, r.EventState&0xFF)
								}
								switch r.EventState & 0xFF {
								case scard.StateEmpty,
									(scard.StateEmpty | scard.StateChanged),
									(scard.StateEmpty | scard.StateExclusive),
									(scard.StateEmpty | scard.StateChanged | scard.StateExclusive):
									// fmt.Printf("state: %X\n", r.EventState&0xFF)
									// fmt.Printf("reader: %q\n", r.Reader)
									if v, ok := app.cardsReader[r.Reader]; ok {
										// if (r.EventState & scard.StateExclusive) != 0x00 {
										fmt.Printf("disconnect state: %X\n", r.EventState&0xFF)
										// }
										v.Disconnect()
										delete(app.cardsReader, r.Reader)
									}
									// if v, ok := app.sessionReader[r.Reader]; ok {
									// 	delete(app.sessionReader, r.Reader)
									// 	delete(app.cardsSession, v)
									// }
									// // default:
									// // 	fmt.Printf("state: %X\n", r.EventState&0xFF)
									// // 	fmt.Printf("reader: %q\n", r.Reader)
								}
							}
							// fmt.Printf("readers state: %+v\n", readers)
						}
					}()
				case sClose, sFatal, sRestart:
					func() {
						app.mux.Lock()
						defer app.mux.Unlock()
						// app.cardsSession = make(map[string]*card.Card)
						for k, v := range app.cardsReader {
							v.Disconnect()
							delete(app.cardsReader, k)
						}
						// app.cardsReader = make(map[string]*card.Card)
						// app.sessionReader = make(map[string]string)
						if app.ctx != nil {
							if ok, err := app.ctx.IsValid(); err != nil || !ok {
								fmt.Printf("error context: %s, success: %v\n", err, ok)
								app.ctx.Release()
							} else {
								app.ctx.Release()
							}
							app.ctx = nil
						}
					}()
					app.fmachine.Event(eOpenCmd)
				}

				// time.Sleep(100 * time.Millisecond)
			}
		}
	}()
	time.Sleep(1 * time.Second)
}
