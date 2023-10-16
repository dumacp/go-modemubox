package main

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/dumacp/go-modemubox"
)

func IpGet(p io.ReadWriter, cid int) (string, string, error) {
	ncid, ip, err := modemubox.PDPcontextReadDynamicParametersGetIP(p)
	if err != nil {
		return "", "", err
	}
	if cid != ncid {
		return "", "", fmt.Errorf("cid: %d, with IP (%s) is not active", ncid, ip)
	}
	ipusb, err := modemubox.GettheUSBIPconfiguration(p, cid)
	if err != nil {
		return "", "", err
	}
	fmt.Println(ipusb)

	return "", "", nil
}

func IPSet(ip, ipusb, ifusb string) error {

	usbdowncmd := fmt.Sprintf("ip link set dev %s down", ifusb)
	testcmd := fmt.Sprintf("ip ad ls dev %s | grep %s | grep %s", ip, ipusb, ifusb)

	usbupcmd := fmt.Sprintf(" ip link set dev %s up", ifusb)
	ipadcmd := fmt.Sprintf("ifconfig usb0:0 %s netmask 255.255.255.255 pointopoint %s up ", ip, ipusb)
	roadcmd := fmt.Sprintf("ip ro ad 0.0.0.0/0 via %s", ipusb)

	if out, err := exec.Command("/bin/sh", "-c", testcmd).Output(); err != nil {
		fmt.Printf("testcmd cmd error, %q, %s\n", out, err)
	} else {
		return nil
	}
	if out, err := exec.Command("/bin/sh", "-c", usbdowncmd).Output(); err != nil {
		return fmt.Errorf("usbdowncmd cmd error, %q, %w", out, err)
	}
	if out, err := exec.Command("/bin/sh", "-c", usbupcmd).Output(); err != nil {
		return fmt.Errorf("usbupcmd cmd error, %q, %w", out, err)
	}
	if out, err := exec.Command("/bin/sh", "-c", ipadcmd).Output(); err != nil {
		return fmt.Errorf("ipadcmd cmd error, %q, %w", out, err)
	}
	if out, err := exec.Command("/bin/sh", "-c", roadcmd).Output(); err != nil {
		return fmt.Errorf("roadcmd cmd error, %q, %w", out, err)
	}

	return nil

}
