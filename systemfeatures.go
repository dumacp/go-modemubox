package modemubox

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type UsbFunction string

const (
	ECM   UsbFunction = "ECM"
	RNDIS UsbFunction = "RNDIS"
)

type UsbProductCategory int

const (
	FairlyBackCompatible     UsbProductCategory = 0
	LowMedium_Throughput     UsbProductCategory = 2
	High_Throughput          UsbProductCategory = 3
	LowMedium_Throughput_SAP UsbProductCategory = 12
	High_Throughput_SAP      UsbProductCategory = 13
)

func InterfaceProfileConf(port io.ReadWriter, upc UsbProductCategory, fun UsbFunction) error {

	cmd := strings.Builder{}
	cmd.WriteString("+UUSBCONF=")
	cmd.WriteString(fmt.Sprintf("%d", upc))
	cmd.WriteString(string(fun))

	res, err := CommandAT(port, cmd.String(), "", 3*time.Second)
	if err != nil {
		return fmt.Errorf("error response: %q", res)
	}

	return nil

}

func GetInterfaceProfileConf(port io.ReadWriter) (UsbProductCategory, UsbFunction, error) {

	cmd := strings.Builder{}
	cmd.WriteString("+UUSBCONF?")

	res, err := sendcommandOneTypeResponse(port, cmd.String(), 1*time.Second)
	if err != nil {
		return 0, "", fmt.Errorf("GetInterfaceProfileConf error: %w", err)
	}

	if len(res) < 2 {
		return 0, "", fmt.Errorf("wring response: %s", res)
	}

	re := regexp.MustCompile(`(\d+),(\w+)$`)
	result := extractParseData(re, res[1])

	if len(result) < 2 {
		return 0, "", fmt.Errorf("invalida data response %s", result)
	}

	upc, _ := strconv.Atoi(result[0])

	return UsbProductCategory(upc), UsbFunction(result[1]), nil

}

func OpenModemConf(port io.ReadWriter, offline bool) error {

	cmd := strings.Builder{}
	cmd.WriteString("+CFUN=")
	if offline {
		cmd.WriteString(fmt.Sprintf("%d", 0))
	} else {
		cmd.WriteString(fmt.Sprintf("%d", 4))
	}

	res, err := CommandAT(port, cmd.String(), "", 3*time.Second)
	if err != nil {
		return fmt.Errorf("error response: %q", res)
	}

	return nil

}

func CloseModemConf(port io.ReadWriter, saveconf bool) error {

	cmd := strings.Builder{}
	cmd.WriteString("+CFUN=")
	if saveconf {
		cmd.WriteString("1,1")
	} else {
		cmd.WriteString("1")
	}

	res, err := CommandAT(port, cmd.String(), "", 3*time.Second)
	if err != nil {
		return fmt.Errorf("error response: %q", res)
	}

	return nil

}
