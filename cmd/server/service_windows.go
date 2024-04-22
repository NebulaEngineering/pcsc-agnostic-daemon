package main

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
)

func (m *myService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	tick := time.NewTicker(500 * time.Millisecond)
	defer tick.Stop()
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		select {
		case <-m.ctx.Done():
			break loop
		case <-tick.C:
			// Aquí colocas la lógica de tu servicio
		case c := <-r:
			switch c.Cmd {
			case svc.Stop, svc.Shutdown:
				m.cancelFunc()
				break loop
			default:
				log.Printf("comando inesperado %v", c)
			}
		}
	}

	changes <- svc.Status{State: svc.StopPending}
	return
}

func runService(name string, isInteractive bool, s *myService) {
	var err error
	if isInteractive {
		err = debug.Run(name, s)
	} else {
		err = svc.Run(name, s)
	}
	if err != nil {
		log.Fatalf("Service %s failed: %v", name, err)
		evtlog := debug.New(name)
		evtlog.Warning(1, fmt.Sprintf("Service %s failed: %v", name, err))
	}
}

func IsAnInteractiveSession() bool {
	ok, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
	}
	return !ok
}
