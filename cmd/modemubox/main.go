package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/tarm/serial"
)

var (
	portpath string
	apns     string
	testip   string
)

func init() {
	flag.StringVar(&portpath, "port", "/dev/tty_modem4g", "path to dev serial")
	flag.StringVar(&apns, "apn", "", "APN name")
	flag.StringVar(&testip, "testip", "8.8.8.8", "test ip (icmp request)")
}

func main() {

	if err := run(); err != nil {
		log.Println(err)
		if err := gpioReset(); err != nil {
			log.Println("error reset \"gsm-reset\": ", err)
		}
	}
}

func run() error {

	flag.Parse()

	c := serial.Config{
		Name:        portpath,
		Baud:        115200,
		ReadTimeout: 1 * time.Second,
	}

	tickPing := time.NewTicker(30 * time.Second)
	defer tickPing.Stop()

	afteInit := time.NewTimer(100 * time.Millisecond)
	defer afteInit.Stop()

	chIPSet := make(chan struct{}, 1)
	chPingTest := make(chan struct{}, 1)
	chContextTest := make(chan struct{}, 1)
	chGpioReset := make(chan struct{}, 1)

	cid := 0

	for {
		select {
		case <-tickPing.C:
			select {
			case chPingTest <- struct{}{}:
			default:
			}
		case <-chPingTest:
			if err := func() error {
				if err := ping(testip, 3, 1, 3); err != nil {
					return fmt.Errorf("error ping: %s", err)
				}
				return nil
			}(); err != nil {
				fmt.Println(err)
				select {
				case chIPSet <- struct{}{}:
				default:
				}
			}
		case <-chIPSet:
			if err := func() error {
				if cid == 0 {
					return fmt.Errorf("cid is 0 (ZERO)")
				}
				p, err := serial.OpenPort(&c)
				if err != nil {
					return err
				}
				defer p.Close()

				ip, ipusb, err := IpGet(p, cid)
				if err != nil {
					fmt.Printf("error IPGet: %s\n", err)
				}
				fmt.Printf("ip: %q, ipusb: %q\n", ip, ipusb)
				if err := IPSet(ip, ipusb, "usb0"); err != nil {
					return fmt.Errorf("error IPSet: %s", err)
				}
				return nil
			}(); err != nil {
				fmt.Println(err)
				select {
				case chContextTest <- struct{}{}:
				default:
				}
			} else {
				select {
				case chPingTest <- struct{}{}:
				default:
				}
			}
		case <-chContextTest:
			if err := func() error {
				p, err := serial.OpenPort(&c)
				if err != nil {
					return err
				}
				defer p.Close()
				if ncid, err := VerifyContext(p, []string{apns, ""}); err != nil {
					return err
				} else {
					cid = ncid
					fmt.Printf("cid: %d\n", cid)
				}
				return nil
			}(); err != nil {
				fmt.Println(err)
				if errors.Is(err, ErrorAT) {
					select {
					case chGpioReset <- struct{}{}:
					default:
					}
				}
			} else {
				select {
				case chPingTest <- struct{}{}:
				default:
				}
			}
		case <-chGpioReset:
			if err := gpioReset(); err != nil {
				return fmt.Errorf("gpio reset error: %s", err)
			}
		case <-afteInit.C:
			select {
			case chContextTest <- struct{}{}:
			default:
			}
		}
	}
}
