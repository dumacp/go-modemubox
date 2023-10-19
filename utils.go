package modemubox

import (
	"fmt"
	"io"
	"regexp"
	"time"
)

func sendcommandOneTypeResponse(port io.ReadWriter, cmd string, timeout time.Duration) ([]string, error) {
	res, err := CommandAT(port, cmd, "", timeout)
	if err != nil {
		return nil, fmt.Errorf("error response: %q", res)
	}

	if len(res) > 0 {
		return extractData(res), nil
	}
	return res, nil
}

func sendcommandOneTypeResponseWithPrefix(port io.ReadWriter, cmd string, timeout time.Duration) (map[string][]string, error) {
	res, err := CommandAT(port, cmd, "", timeout)
	if err != nil {
		return nil, fmt.Errorf("error response: %q", res)
	}

	if len(res) > 0 {
		return extractDataWithPrefix(res), nil
	}
	return nil, nil
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
	fmt.Println("extractData response: ", results)
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

func extractParseData(re *regexp.Regexp, s string) []string {

	match := re.FindStringSubmatch(s)
	if len(match) > 1 {
		return match[1:]
	}

	return nil
}
