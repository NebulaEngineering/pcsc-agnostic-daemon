package main

import (
	"fmt"
	"log"
	"time"

	"gitlab.com/nebulaeng/fleet/pcscrest/internal/app"
	"gitlab.com/nebulaeng/fleet/pcscrest/internal/pcsc/card"
)

func main() {

	rds, err := app.Instance().ListReaders()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("readers: %v\n", rds)

	t1 := time.NewTicker(3 * time.Second)
	defer t1.Stop()
	var c *card.Card
	for range t1.C {

		if err := func() error {
			if c == nil {
				c, err = app.Instance().ConnectCardInReader(rds[0])
				if err != nil {
					return err
				}
			}
			status, err := c.Status()
			if err != nil {
				return err
			}

			fmt.Printf("card Satus: %v\n", status)

			atr, err := c.Atr()
			if err != nil {
				return err
			}

			fmt.Printf("card ATR: %X\n", atr)
			return nil
		}(); err != nil {
			if c != nil {
				c.Disconnect()
				c = nil
			}
			log.Printf("main error: %s", err)
		}
	}
}
