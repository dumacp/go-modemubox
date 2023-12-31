package modemubox

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

func CommandAT(port io.ReadWriter, cmd, arg string, timeout time.Duration) ([]string, error) {

	cmd = strings.ToUpper(cmd)
	cmdLine := strings.Builder{}

	if len(cmd) > 1 && !strings.HasPrefix(cmd, "AT") {
		cmdLine.WriteString("AT" + cmd)
	} else {
		cmdLine.WriteString(cmd)
	}
	if len(arg) > 0 {
		cmdLine.WriteString("=" + arg + "\r")
	} else {
		cmdLine.WriteString("\r")
	}

	// Leer líneas de respuesta hasta que llegue "OK" o "ERROR" o se alcance el timeout
	ch := make(chan string)
	errc := make(chan error)
	go func(cmd string) {
		defer close(ch)

		after := time.NewTimer(timeout)
		defer after.Stop()

		reader := bufio.NewReader(port)

		withResponse := false
		for {

			select {
			case <-after.C:
				errc <- fmt.Errorf("timeout")
				return
			default:
				line, err := reader.ReadString('\r')
				if err != nil {
					if withResponse && errors.Is(err, io.EOF) {
						continue
					}
					errc <- err
					return
				}
				if !withResponse && strings.HasPrefix(line, cmd) {
					withResponse = true
				}

				ch <- strings.TrimSpace(line)

				// Verificar si la línea contiene "OK" o "ERROR"
				if strings.Contains(line, "OK") {
					return
				}
				if strings.Contains(line, "ERROR") {
					errc <- fmt.Errorf("error in response")
					return
				}
			}

		}

	}(cmdLine.String())

	// Enviar el comando al dispositivo

	_, err := port.Write([]byte(cmdLine.String()))
	if err != nil {
		return nil, err
	}

	lines := make([]string, 0)

break_for:
	for {
		select {
		case v, ok := <-ch:
			if !ok {
				break break_for
			}
			lines = append(lines, v)
		case err, ok := <-errc:
			if !ok {
				break break_for
			}
			return lines, err
		}
	}

	fmt.Println("commandAT response: ", lines)

	return lines, nil
}
