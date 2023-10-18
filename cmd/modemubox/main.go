package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/dumacp/go-modemubox"
	"github.com/tarm/serial"
)

var (
	portpath   string
	apns       StringSlice
	testip     string
	iptestenvs []string
)

type StringSlice []string

func (s *StringSlice) String() string {
	return strings.Join(*s, ",")
}

func (s *StringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func (s *StringSlice) Value() []string {
	sl := make([]string, 0)
	for _, v := range *s {
		if len(v) == 0 {
			continue
		}
		sl = append(sl, v)
	}
	sl = append(sl, "")
	return sl
}

func init() {
	flag.StringVar(&portpath, "port", "/dev/tty_modem4g", "path to dev serial")
	flag.Var(&apns, "apn", "APN name")
	flag.StringVar(&testip, "testip", "8.8.8.8", "test ip (icmp request)")
}

func main() {

	flag.Parse()

	for i, item := range apns {
		fmt.Printf("apn %d: %s\n", i+1, item)
	}

	apnenv, _ := getAPN()
	apnenvs := strings.Split(apnenv, ",")
	if len(apnenvs) > 0 && len(apnenvs[0]) > 0 {
		apns = apnenvs
	}

	for i, item := range apns {
		fmt.Printf("apn from Environment %d: %s\n", i+1, item)
	}

	iptestenv, _ := getTestIP()
	iptestenvs = strings.Split(iptestenv, ",")
	if len(iptestenvs) <= 0 || len(iptestenvs[0]) <= 0 {
		iptestenvs = []string{testip}
	}

	for i, item := range iptestenvs {
		fmt.Printf("iptest from Environment %d: %s\n", i+1, item)
	}

	if err := run(); err != nil {
		log.Println(err)
		if err := gpioReset(); err != nil {
			log.Println("error reset \"gsm-reset\": ", err)
		}
	}
}

func run() error {

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
	chGpioPower := make(chan struct{}, 1)

	cid := 0
	const MaxError = 4
	countError := 5

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	chKmesg, err := modemubox.TailKmesg(ctx)
	if err != nil {
		fmt.Println("error tail kmesg: ", err)
	}
	defer close(chKmesg)

	for {
		select {
		case v, ok := <-chKmesg:
			if !ok {
				return fmt.Errorf("error tail kmesg: %s", err)
			}
			if strings.Contains(v, "cdc_ether") {
				fmt.Println("kmesg event: ", v)
				if strings.Contains(v, "unregister") {
					select {
					case chPingTest <- struct{}{}:
					default:
					}
				}
			}
		case <-tickPing.C:
			select {
			case chPingTest <- struct{}{}:
			default:
			}
		case <-chPingTest:
			if err := func() error {
				if err := ping(iptestenvs, 3, 1, 3); err != nil {
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

				if ncid, err := VerifyContext(p, apns.Value()); err != nil {
					return err
				} else {
					cid = ncid
					fmt.Printf("cid: %d\n", cid)
				}
				return nil
			}(); err != nil {
				var patErr *os.PathError
				fmt.Println(err)
				if errors.Is(err, ErrorAT) ||
					errors.As(err, &patErr) {
					countError++
					if countError <= 1 {
						time.Sleep(3 * time.Second)
						select {
						case chPingTest <- struct{}{}:
						default:
						}
					} else if countError > MaxError {
						countError = 0
						select {
						case chGpioPower <- struct{}{}:
						default:
						}
					} else {
						countError = 0
						select {
						case chGpioReset <- struct{}{}:
						default:
						}
					}
				}
			} else {
				select {
				case chPingTest <- struct{}{}:
				default:
				}
			}
		case <-chGpioReset:
			fmt.Println("gpio reset")
			if err := gpioReset(); err != nil {
				return fmt.Errorf("gpio reset error: %s", err)
			}
		case <-chGpioPower:
			fmt.Println("gpio power")
			if err := gpioPower(); err != nil {
				return fmt.Errorf("gpio power error: %s", err)
			}
		case <-afteInit.C:
			select {
			case chContextTest <- struct{}{}:
			default:
			}
		}
	}
}
