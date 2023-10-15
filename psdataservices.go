package modemubox

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type PDPType string

const (
	IP     PDPType = "IP"     // Internet Protocol (IETF STD 5)
	NONIP  PDPType = "NONIP"  // Non IP
	IPV4V6 PDPType = "IPV4V6" //virtual <PDP_type> introduced to handle dual IP stack UE capability
)

func PDPcontextActivate(port io.ReadWriteCloser, cid int, active bool) error {
	cmd := strings.Builder{}
	cmd.WriteString("AT+CGACT=")
	cmd.WriteString(fmt.Sprintf("%d,%d", func() int {
		if active {
			return 1
		}
		return 0
	}(), cid))
	if res, err := CommandAT(port, cmd.String(), "", 5*time.Second); err != nil {
		return fmt.Errorf("error response: %q", res)
	}
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

	cmd.WriteString(fmt.Sprintf("%q,%d,%d,%d,%d,%d,%d", ip, d_comp, h_comp, ipv4Alloc, emer_ind, req_type, P_CSCF_discovery))
	if res, err := CommandAT(port, cmd.String(), "", 1*time.Second); err != nil {
		return fmt.Errorf("error response: %q", res)
	}

	return nil
}

func PDPcontextDefinitionShort(port io.ReadWriteCloser, cid int, pdptype PDPType, apn string) error {
	return PDPcontextDefinition(port, cid, pdptype, apn, "", 0, 0, 0, 0, 0, 0)
}

type PDPcontextParameters struct {
	Cid                         int
	BearerID                    int
	APN                         string
	LocalAddrAndSubnet          string
	QgwAddr                     string
	DNSPrimAddr                 string
	DNSSecAddr                  string
	PCSCFPrimAddr               string
	PCSCFSecAddr                string
	IMCNSignallingFlag          int
	LIPAIndication              int
	IPv4MTU                     int
	WLANOffload                 int
	LocalAddrInd                int
	NonIPMTU                    int
	ServingPLMNRateControlValue int
}

func ParseToPDPcontextParameters(input string) PDPcontextParameters {

	p := PDPcontextParameters{}

	records := strings.FieldsFunc(input, func(r rune) bool { return r == ',' })

	if len(records) <= 0 {
		return p
	}
	cid, err := strconv.Atoi(records[0])
	if err != nil {
		return p
	}
	p.Cid = cid
	if len(records) <= 0 {
		return p
	}
	bearerID, err := strconv.Atoi(records[2])
	if err != nil {
		return p
	}
	p.BearerID = bearerID

	if len(records) <= 0 {
		return p
	}
	p.APN = records[1]

	if len(records) <= 0 {
		return p
	}
	localAddrAndSubnetpn := records[3]
	if len(records) <= 0 {
		return p
	}
	p.LocalAddrAndSubnet = localAddrAndSubnetpn
	qgwAddr := records[4]
	if len(records) <= 0 {
		return p
	}
	p.QgwAddr = qgwAddr
	dNSPrimAddr := records[5]
	if len(records) <= 0 {
		return p
	}
	p.DNSPrimAddr = dNSPrimAddr
	dNSSecAddr := records[6]
	if len(records) <= 0 {
		return p
	}
	p.DNSSecAddr = dNSSecAddr
	pCSCFPrimAddr := records[7]
	if len(records) <= 0 {
		return p
	}
	p.PCSCFPrimAddr = pCSCFPrimAddr
	pCSCFSecAddr := records[8]
	if len(records) <= 0 {
		return p
	}
	p.PCSCFSecAddr = pCSCFSecAddr
	if len(records) <= 9 {
		return p
	}
	imCNSignallingFlag, err := strconv.Atoi(records[9])
	if err != nil {
		return p
	}
	p.IMCNSignallingFlag = imCNSignallingFlag
	if len(records) <= 10 {
		return p
	}
	lIPAIndication, err := strconv.Atoi(records[10])
	if err != nil {
		return p
	}
	p.LIPAIndication = lIPAIndication
	if len(records) <= 11 {
		return p
	}
	iPv4MTU, err := strconv.Atoi(records[9])
	if err != nil {
		return p
	}
	p.IPv4MTU = iPv4MTU
	if len(records) <= 12 {
		return p
	}
	wLANOffload, err := strconv.Atoi(records[9])
	if err != nil {
		return p
	}
	p.WLANOffload = wLANOffload
	if len(records) <= 13 {
		return p
	}
	localAddrInd, err := strconv.Atoi(records[9])
	if err != nil {
		return p
	}
	p.LocalAddrInd = localAddrInd
	if len(records) <= 14 {
		return p
	}
	nonIPMTU, err := strconv.Atoi(records[9])
	if err != nil {
		return p
	}
	p.NonIPMTU = nonIPMTU
	if len(records) <= 15 {
		return p
	}
	servingPLMNRateControlValue, err := strconv.Atoi(records[9])
	if err != nil {
		return p
	}
	p.ServingPLMNRateControlValue = servingPLMNRateControlValue

	return p
}

func PDPcontextReadDynamicParametersGetIP(port io.ReadWriteCloser) (int, string, error) {
	cmd := strings.Builder{}
	cmd.WriteString("+CGCONTRDP")
	res, err := CommandAT(port, cmd.String(), "", 3*time.Second)
	if err != nil {
		return 0, "", fmt.Errorf("error response: %q", res)
	}

	return getCidAndIP(res)
}

func getCidAndIP(s []string) (int, string, error) {
	re := regexp.MustCompile(`(\d{1,3}\.){3}\d{1,3}`)

	for _, v := range s {
		match := re.FindString(v)
		if len(match) > 0 {
			reCid := regexp.MustCompile(`\+CGCONTRDP: (\d+)`)
			matchCid := reCid.FindStringSubmatch(v)
			if len(matchCid) > 1 {
				cid, err := strconv.Atoi(matchCid[1])
				if err != nil {
					return 0, "", err
				}
				return cid, match, nil
			}
		}
	}

	return 0, "", fmt.Errorf("ip not found")
}
