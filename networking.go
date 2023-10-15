package modemubox

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

func GettheUSBIPconfiguration(port io.ReadWriteCloser, cid int) (string, error) {

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
