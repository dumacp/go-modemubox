package modemubox

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"
)

func CommandAT(cmd, arg string, port io.ReadWriteCloser, timeout time.Duration) ([]string, error) {

	// Leer líneas de respuesta hasta que llegue "OK" o "ERROR" o se alcance el timeout
	ch := make(chan string)
	errc := make(chan error)
	go func() {
		defer close(ch)

		after := time.NewTimer(timeout)
		defer after.Stop()

		reader := bufio.NewReader(port)
		for {

			select {
			case <-after.C:
				errc <- fmt.Errorf("timeout")
				return
			default:
				line, err := reader.ReadString('\r')
				if err != nil {
					errc <- err
					return
				}

				ch <- strings.TrimSpace(line)

				// Verificar si la línea contiene "OK" o "ERROR"
				if strings.Contains(line, "OK") || strings.Contains(line, "ERROR") {
					return
				}
			}

		}

	}()

	// Enviar el comando al dispositivo

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

	return lines, nil
}
