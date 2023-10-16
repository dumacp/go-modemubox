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

func VerifyContext(p *serial.Port, apns []string) (int, error) {

	if _, err := modemubox.CommandAT(p, "AT", "", 1*time.Second); err != nil {
		return 0, ErrorAT
	}

	at, pt, err := modemubox.GetRadioAccessTechnologySelection(p)
	if err != nil {
		return 0, fmt.Errorf("getRadioAccessTechnologySelection error: %w ", err)
	}

	if at != modemubox.GSM_UMTS_dual_mode || pt != modemubox.GSM_GPRS_eGPRS {
		modemubox.RadioAccessTechnologySelection(p, modemubox.GSM_UMTS_dual_mode, modemubox.GSM_GPRS_eGPRS)
		time.Sleep(10 * time.Second)
	}

	currents, err := getPdp(p)
	if err != nil {
		return 0, fmt.Errorf("get currents APN error: %w", err)
	}

	cidApn := parseapn(currents)
	// for k, v := range apns {

	// 	if err := checkPdp(p, k+1, v, currents); err != nil {
	// 		return 0, fmt.Errorf("set APN error: %w", err)
	// 	}
	// }

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

	currentApn := cidApn[cid]
	fmt.Printf("current CGPCONT: %q\n", cidApn[cid])

	for _, v := range apns {
		if len(v) > 0 && strings.Contains(currentApn, v) {
			return cid, nil
		}
	}

	// if cid == 0 {
	var errx error
	for k, v := range apns {
		if err := checkPdp(p, k+1, v, cidApn); err != nil {
			errx = fmt.Errorf("PDPcontextDefinition error: %w", err)
			continue
		}
		if err := modemubox.PDPcontextActivate(p, k+1, true); err != nil {
			errx = fmt.Errorf("PDPcontextActivate error: %w", err)
			continue
		}
		cid = k + 1
		errx = nil
		break
	}
	if errx != nil {
		return 0, errx
	}
	// }

	if apn, ok := cidApn[cid]; ok {
		for _, v := range apns {
			if len(v) == 0 {
				if len(apn) == 0 {
					return cid, nil
				}
				continue
			}
			if strings.Contains(apn, v) {
				return cid, nil
			}
		}
	}
	// fmt.Printf("current CGPCONT: %q\n", cidApn[cid])

	// var errx error
	// for k, v := range apns {
	// 	if err := checkPdp(p, k+1, v, cidApn); err != nil {
	// 		errx = fmt.Errorf("PDPcontextDefinition error: %w", err)
	// 		continue
	// 	}
	// 	if err := modemubox.PDPcontextActivate(p, k+1, true); err != nil {
	// 		errx = fmt.Errorf("PDPcontextActivate error: %w", err)
	// 		continue
	// 	}
	// 	errx = nil
	// 	break
	// }
	// if errx != nil {
	// 	return 0, errx
	// }

	fmt.Printf("current set CGPCONT: %q\n", cidApn[cid])

	return cid, nil
}

func getPdp(port io.ReadWriter) ([]string, error) {

	currents, err := modemubox.GetPDPcontextDefinition(port)
	if err != nil {
		return nil, err
	}
	fmt.Printf("CGDCONT currents: %q\n", currents)
	return currents, nil
}

func checkPdp(port io.ReadWriter, cid int, apn string, currentapns map[int]string) error {

	for _, v := range currentapns {
		if len(apn) == 0 {
			if len(v) == 0 {
				return nil
			}
			continue
		}
		if strings.Contains(v, apn) {
			return nil
		}
	}
	return modemubox.PDPcontextDefinitionShort(port, cid, modemubox.IP, apn)
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

func parseapn(res []string) map[int]string {
	re := regexp.MustCompile(`^(\d+),\"\w+\",\"([[:word:]\.-]*)\",`)
	results := make(map[int]string, 0)
	for _, s := range res {
		match := re.FindStringSubmatch(s)
		if len(match) > 2 {
			key, _ := strconv.Atoi(match[1])
			value := match[2]
			results[key] = value
		}
	}
	return results
}
