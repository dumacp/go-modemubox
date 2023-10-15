package modemubox

import (
	"fmt"
	"io"
	"strings"
	"time"
)

func sendcommandResponse(port io.ReadWriteCloser, cmd string, timeout time.Duration) ([]string, error) {
	res, err := CommandAT(port, cmd, "", timeout)
	if err != nil {
		return nil, fmt.Errorf("error response: %q", res)
	}

	result := make([]string, 0)
	for _, v := range res {
		if strings.Contains(v, cmd) {
			continue
		}
		if len(v) == 2 && strings.Contains(v, "OK") {
			continue
		}
		if len(v) == 0 {
			continue
		}
		result = append(result, v)
	}
	return result, nil
}
