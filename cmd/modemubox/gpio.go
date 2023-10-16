package main

import (
	"io"
	"os"
	"time"
)

func gpioReset() error {

	fReset, err := os.OpenFile("/sys/class/leds/gsm-rst/brightness", os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	if _, err := fReset.Seek(0, io.SeekStart); err != nil {
		return err
	}

	if _, err := fReset.Write([]byte("1")); err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	if _, err := fReset.Write([]byte("0")); err != nil {
		return err
	}
	time.Sleep(10 * time.Second)

	return nil
}

func gpioPower() error {

	fReset, err := os.OpenFile("/sys/class/leds/gsm-pwr/brightness", os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	if _, err := fReset.Seek(0, io.SeekStart); err != nil {
		return err
	}

	if _, err := fReset.Write([]byte("1")); err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	if _, err := fReset.Write([]byte("0")); err != nil {
		return err
	}

	return nil
}
