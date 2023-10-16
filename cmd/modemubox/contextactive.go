package main

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dumacp/go-modemubox"
	"github.com/tarm/serial"
)

func VerifyContext(p *serial.Port) (int, error) {

	if _, err := modemubox.CommandAT(p, "AT", "", 1*time.Second); err != nil {
		return 0, ErrorAT
	}

	if err := checkPdp(p, apn); err != nil {
		return 0, fmt.Errorf("set APN error: %w", err)
	}

	at, pt, err := modemubox.GetRadioAccessTechnologySelection(p)
	if err != nil {
		return 0, fmt.Errorf("getRadioAccessTechnologySelection error: %w ", err)
	}

	if at != modemubox.GSM_UMTS_LTE_tri_mode || pt != modemubox.LTE {
		modemubox.RadioAccessTechnologySelection(p, modemubox.GSM_UMTS_LTE_tri_mode, modemubox.LTE)
		time.Sleep(10 * time.Second)
	}

	ctxs, err := contextActives(p)
	if err != nil {
		return 0, fmt.Errorf("getContextActive error: %w (%q)", err, ctxs)
	}

	cid := 0
	for k, v := range ctxs {
		if v != 0 {
			cid = k
			break
		}
	}
	if cid == 0 {
		if err := modemubox.PDPcontextActivate(p, 1, true); err != nil {
			return 0, fmt.Errorf("PDPcontextActivate error: %w", err)
		}
		cid = 1
	}
	return cid, nil
}

func checkPdp(port io.ReadWriter, apn string) error {

	currents, err := modemubox.GetPDPcontextDefinition(port)
	if err != nil {
		return err
	}

	for _, v := range currents {
		if strings.Contains(v, apn) {
			return nil
		}
	}
	return modemubox.PDPcontextDefinitionShort(port, 1, modemubox.IP, apn)
}

func contextActives(port io.ReadWriter) (map[int]int, error) {

	res, err := modemubox.GetPDPcontextActivates(port)
	if err != nil {
		return nil, err
	}
	results := parsecontext(res)
	return results, nil
}

func parsecontext(res []string) map[int]int {
	re := regexp.MustCompile(`(\d+),(\d+)$`)
	results := make(map[int]int, 0)
	for _, s := range res {
		match := re.FindStringSubmatch(s)
		if len(match) > 2 {
			key, _ := strconv.Atoi(match[1])
			value, _ := strconv.Atoi(match[2])
			results[key] = value
		}
	}
	return results
}
