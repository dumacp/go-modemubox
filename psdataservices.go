package modemubox

import (
	"fmt"
	"io"
	"strings"
	"time"
)

type PDPType string

const (
	IP     PDPType = "IP"     // Internet Protocol (IETF STD 5)
	NONIP  PDPType = "NONIP"  // Non IP
	IPV4V6 PDPType = "IPV4V6" //virtual <PDP_type> introduced to handle dual IP stack UE capability
)

func PDPcontextActivate(active bool) error {
	return nil
}

func PDPcontextDefinition(port io.ReadWriteCloser, cid int, pdptype PDPType, apn, ip string, d_comp, h_comp, ipv4Alloc, emer_ind, req_type, P_CSCF_discovery int) error {

	cmd := strings.Builder{}
	cmd.WriteString("+CGDCONT=")
	cmd.WriteString(fmt.Sprintf("%d,%q,%q", cid, pdptype, apn))
	if (len(ip) <= 0) && d_comp == 0 && h_comp == 0 && ipv4Alloc == 0 && emer_ind == 0 && req_type == 0 && P_CSCF_discovery == 0 {
		if res, err := CommandAT(port, cmd.String(), "", 1*time.Second); err != nil {
			return fmt.Errorf("error response: %q", res)
		}
	}

	return nil
}

func PDPcontextDefinitionShort(port io.ReadWriteCloser, cid int, pdptype PDPType, apn string) error {
	return PDPcontextDefinition(port, cid, pdptype, apn, "", 0, 0, 0, 0, 0, 0)
}
