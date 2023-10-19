package modemubox

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func GetBootModeConfiguration(port io.ReadWriter) (int, error) {

	cmd := strings.Builder{}
	cmd.WriteString("+UBMCONF?")
	res, err := sendcommandOneTypeResponseWithPrefix(port, cmd.String(), 1*time.Second)
	if err != nil {
		return 0, fmt.Errorf("GetInterfaceProfileConf error: %w", err)
	}

	fmt.Println("response: ", res)

	mode := 0
	for k, v := range res {
		if strings.HasPrefix(k, "UBMCONF") {
			var err error
			if len(v) <= 0 {
				return 0, fmt.Errorf("wrong response: %s", res)
			}
			mode, err = strconv.Atoi(v[0])
			if err != nil {
				return 0, fmt.Errorf("wrong response: %s", res)
			}
			break
		}

	}

	return mode, nil
}

func BootModeConfiguration(port io.ReadWriter, mode int) error {

	cmd := strings.Builder{}
	cmd.WriteString("+UBMCONF=")
	cmd.WriteString(fmt.Sprintf("%d", mode))
	res, err := CommandAT(port, cmd.String(), "", 3*time.Second)
	if err != nil {
		return fmt.Errorf("error response: %q", res)
	}

	return nil

}

func GettheUSBIPconfiguration(port io.ReadWriter, cid int) (string, error) {

	cmd := strings.Builder{}
	cmd.WriteString("+UIPADDR=")
	cmd.WriteString(fmt.Sprintf("%d", cid))
	res, err := CommandAT(port, cmd.String(), "", 3*time.Second)
	if err != nil {
		return "", fmt.Errorf("error response: %q", res)
	}

	return getIP(res)
}

func getIP(s []string) (string, error) {
	re := regexp.MustCompile(`(\d{1,3}\.){3}\d{1,3}`)

	for _, v := range s {
		match := re.FindString(v)
		if len(match) > 0 {
			return match, nil
		}
	}
	return "", fmt.Errorf("ip not found")
}
