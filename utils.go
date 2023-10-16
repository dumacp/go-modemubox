package modemubox

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

func sendcommandOneTypeResponse(port io.ReadWriter, cmd string, timeout time.Duration) ([]string, error) {
	res, err := CommandAT(port, cmd, "", timeout)
	if err != nil {
		return nil, fmt.Errorf("error response: %q", res)
	}

	result := make([]string, 0)
	for _, v := range res {
		if strings.HasSuffix(v, cmd) {
			continue
		}
		if len(v) == 2 && strings.Contains(v, "OK") {
			continue
		}
		if len(v) == 0 {
			continue
		}
	}
	if len(result) > 0 {
		return extractData(result), nil
	}
	return result, nil
}

func extractData(data []string) []string {
	re := regexp.MustCompile(`^\+(\w+): (.+)$`)

	results := make([]string, 0)
	for _, s := range data {
		match := re.FindStringSubmatch(s)
		if len(match) > 2 {
			value := match[2]
			results = append(results, value)
		}
	}
	return results
}

func extractDataWithPrefix(data []string) map[string][]string {
	re := regexp.MustCompile(`^\+(\w+): (.+)$`)

	results := make(map[string][]string)
	for _, s := range data {
		match := re.FindStringSubmatch(s)
		if len(match) > 2 {
			key := match[1]
			value := match[2]
			results[key] = append(results[key], value)
		}
	}
	return results
}
