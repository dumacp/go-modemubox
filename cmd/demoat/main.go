package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/dumacp/go-modemubox"
	"github.com/tarm/serial"
)

var (
	portpath string
	cmd      string
)

func init() {
	flag.StringVar(&portpath, "port", "/dev/tty_modem4g", "path to dev serial")
	flag.StringVar(&cmd, "cmd", "AT", "path to dev serial")
}

func main() {

	flag.Parse()

	c := serial.Config{
		Name:        portpath,
		Baud:        115200,
		ReadTimeout: 1 * time.Second,
	}

	p, err := serial.OpenPort(&c)
	if err != nil {
		log.Fatalln(err)
	}

	result, err := modemubox.CommandAT(cmd, "", p, 3*time.Second)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("respoonse ", result)

}
