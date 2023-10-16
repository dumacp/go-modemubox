package modemubox

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type AccessTecnology int

const (
	GSM_GPRS_eGPRS_single_mode AccessTecnology = 0 // GSM / GPRS / eGPRS (single mode)
	GSM_UMTS_dual_mode         AccessTecnology = 1 // GSM / UMTS (dual mode)
	UMTS_single_mode           AccessTecnology = 2 // UMTS (single mode)
	LTE_single_mode            AccessTecnology = 3 // LTE (single mode)
	GSM_UMTS_LTE_tri_mode      AccessTecnology = 4 // GSM / UMTS / LTE (tri mode)
	GSM_LTE_dual_mode          AccessTecnology = 5 // GSM / LTE (dual mode)
	UMTS_LTE_dual_mode         AccessTecnology = 6 // UMTS / LTE (dual mode)
)

type PreferedAccessTecnology int

const (
	GSM_GPRS_eGPRS PreferedAccessTecnology = 0 // GSM / GPRS / eGPRS (single mode)
	UTRAN          PreferedAccessTecnology = 2 // GSM / UMTS (dual mode)
	LTE            PreferedAccessTecnology = 3 // UMTS (single mode)
)

/**
• 0: GSM / GPRS / eGPRS (single mode)
• 1: GSM / UMTS (dual mode)
• 2: UMTS (single mode)
• 3: LTE (single mode)
• 4: GSM / UMTS / LTE (tri mode)
• 5: GSM / LTE (dual mode)
• 6: UMTS / LTE (dual mode)


• 0: GSM / GPRS / eGPRS
• 2: UTRAN
• 3: LTE
**/

func RadioAccessTechnologySelection(port io.ReadWriter, selectAt AccessTecnology, preferedAt PreferedAccessTecnology) error {
	cmd := strings.Builder{}
	cmd.WriteString("AT+URAT=")
	cmd.WriteString(fmt.Sprintf("%d,%d", selectAt, preferedAt))
	if res, err := CommandAT(port, "+CFUN=4", "", 1*time.Second); err != nil {
		return fmt.Errorf("error response: %q", res)
	}
	if res, err := CommandAT(port, cmd.String(), "", 5*time.Second); err != nil {
		return fmt.Errorf("error response: %q", res)
	}
	if res, err := CommandAT(port, "+CFUN=1", "", 2*time.Second); err != nil {
		return fmt.Errorf("error response: %q", res)
	}
	return nil
}

func GetRadioAccessTechnologySelection(port io.ReadWriter) (AccessTecnology, PreferedAccessTecnology, error) {

	cmd := strings.Builder{}
	cmd.WriteString("+URAT?")

	res, err := sendcommandOneTypeResponse(port, cmd.String(), 1*time.Second)
	if err != nil {
		return 0, 0, fmt.Errorf("getRadioAccessTechnologySelection error: %w", err)
	}

	at, pt := parseaccessTechnology(res)

	return AccessTecnology(at), PreferedAccessTecnology(pt), nil

}

func parseaccessTechnology(res []string) (int, int) {
	re := regexp.MustCompile(`(\d+),(\d+)$`)
	for _, s := range res {
		match := re.FindStringSubmatch(s)
		if len(match) > 2 {
			key, _ := strconv.Atoi(match[1])
			value, _ := strconv.Atoi(match[2])
			return key, value
		}
	}
	return 0, 0
}
