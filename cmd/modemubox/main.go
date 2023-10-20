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
	atecn      int
	ptecn      int
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
	flag.IntVar(&atecn, "AT", int(modemubox.GSM_UMTS_LTE_tri_mode), fmt.Sprintf("AccessTecnologyValues (values: %s)", modemubox.AccessTecnologyValues()))
	flag.IntVar(&ptecn, "PT", int(modemubox.LTE), fmt.Sprintf("PreferedAccessTecnologyValues (values: %s)", modemubox.PreferedAccessTecnologyValues()))
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

	c := serial.Config{
		Name:        portpath,
		Baud:        115200,
		ReadTimeout: 1 * time.Second,
	}

	if err := initial(c); err != nil {
		log.Println(err)
		time.Sleep(1 * time.Second)

		if err := gpioPower(); err != nil {
			log.Println("error reset \"gsm-reset\": ", err)
		}
		return
	}

	if err := run(c); err != nil {
		log.Println(err)
		if err := gpioPower(); err != nil {
			log.Println("error reset \"gsm-reset\": ", err)
		}
	}
}

func initial(c serial.Config) error {

	p, err := serial.OpenPort(&c)
	if err != nil {
		return err
	}
	defer p.Close()

	mode, err := modemubox.GetBootModeConfiguration(p)
	if err != nil {
		return err
	}

	at, pt, err := modemubox.GetRadioAccessTechnologySelection(p)
	if err != nil {
		return err
	}

	upc, uf, err := modemubox.GetInterfaceProfileConf(p)
	if err != nil {
		return err
	}

	if mode != 2 ||
		at != modemubox.AccessTecnology(atecn) || pt != modemubox.PreferedAccessTecnology(ptecn) ||
		upc != modemubox.LowMedium_Throughput || uf != modemubox.ECM {
		if err := modemubox.OpenModemConf(p, true); err != nil {
			return err
		}
		defer modemubox.CloseModemConf(p, true)
	}

	if mode != 2 {
		if err := modemubox.BootModeConfiguration(p, 2); err != nil {
			return err
		}
	}

	if at != modemubox.AccessTecnology(atecn) || pt != modemubox.PreferedAccessTecnology(ptecn) {
		if err := modemubox.RadioAccessTechnologySelection(p, modemubox.AccessTecnology(atecn), modemubox.PreferedAccessTecnology(ptecn)); err != nil {
			return err
		}
	}

	if upc != modemubox.LowMedium_Throughput || uf != modemubox.ECM {
		if err := modemubox.InterfaceProfileConf(p, modemubox.LowMedium_Throughput, modemubox.ECM); err != nil {
			return err
		}
	}

	return nil
}

func run(c serial.Config) error {

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
	const MaxError = 10
	countError := 0
	const MaxPingError = 30
	countPingError := 0
	lastReset := time.Now().Add(-30 * time.Second)
	lastErrATcommand := time.Now().Add(-300 * time.Second)

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	chKmesg, err := modemubox.TailKmesg(ctx)
	if err != nil {
		fmt.Println("error tail kmesg: ", err)
	}
	defer close(chKmesg)

	ip := ""
	ipusb := ""
	for {
		if time.Since(lastReset) < 30*time.Second {
			time.Sleep(1 * time.Second)
			continue
		}
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
					if len(ip) > 0 && len(ipusb) > 0 {
						fmt.Printf("ip: %q, ipusb: %q\n", ip, ipusb)
						if err := IPSet(ip, ipusb, "usb0"); err != nil {
							return fmt.Errorf("error IPSet: %s", err)
						}
						if err := ping(iptestenvs, 3, 1, 3); err != nil {
							return fmt.Errorf("error ping: %s", err)
						}
					} else {
						return err
					}
				}
				return nil
			}(); err != nil {
				fmt.Println(err)
				countPingError++
				if countPingError > MaxPingError {
					select {
					case chGpioPower <- struct{}{}:
						countPingError = 0
					default:
					}
				} else {
					select {
					case chIPSet <- struct{}{}:
					default:
					}
				}
			} else {
				countPingError = 0
			}
		case <-chIPSet:
			// fmt.Println(1)
			if time.Since(lastErrATcommand) < 60*time.Second {
				break
			}
			// fmt.Println(1)
			if err := func() error {
				if cid == 0 {
					return fmt.Errorf("cid is 0 (ZERO)")
				}
				p, err := serial.OpenPort(&c)
				if err != nil {
					return err
				}
				defer p.Close()

				// var err error
				ip, ipusb, err = IpGet(p, cid)
				if err != nil {
					return err
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
			if time.Since(lastErrATcommand) < 60*time.Second {
				break
			}
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
					lastErrATcommand = time.Now()
					countError++
					/** if countError <= MaxError/2 {
						time.Sleep(3 * time.Second)
						select {
						case chPingTest <- struct{}{}:
						default:
						}
					} else /**/if countError > MaxError {
						time.Sleep(3 * time.Second)
						select {
						case chGpioPower <- struct{}{}:
							countError = 0
						default:
						}
					} else {
						// countError = 0
						time.Sleep(3 * time.Second)
						select {
						// case chGpioReset <- struct{}{}:
						case chPingTest <- struct{}{}:
						default:
						}
					}
				}
			} else {
				countError = 0
				select {
				case chPingTest <- struct{}{}:
				default:
				}
			}
		case <-chGpioReset:
			if time.Since(lastReset) < 60*time.Second {
				break
			}
			lastReset = time.Now()
			fmt.Println("gpio reset")
			if err := gpioReset(); err != nil {
				return fmt.Errorf("gpio reset error: %s", err)
			}
		case <-chGpioPower:
			if time.Since(lastReset) < 300*time.Second {
				break
			}
			lastReset = time.Now()
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
