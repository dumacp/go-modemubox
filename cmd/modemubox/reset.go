package main

import (
	"time"

	"github.com/dumacp/go-modemubox"
	"github.com/tarm/serial"
)

func OfflineOnline(c *serial.Config) error {

	p, err := serial.OpenPort(c)
	if err != nil {
		return err
	}
	defer p.Close()

	if _, err := modemubox.CommandAT(p, "+CFUN=4", "", 1*time.Second); err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	if _, err := modemubox.CommandAT(p, "+CFUN=1", "", 2*time.Second); err != nil {
		return err
	}
	time.Sleep(3 * time.Second)

	return nil
}
