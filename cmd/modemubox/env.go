package main

import (
	"fmt"
	"os"
)

func getTestIP() (string, error) {
	testIP := os.Getenv("TEST_IP")
	if len(testIP) <= 0 {
		return "", fmt.Errorf("TEST_IP not found")
	}
	return testIP, nil
}

func getAPN() (string, error) {
	apn := os.Getenv("APN")
	if len(apn) <= 0 {
		return "", fmt.Errorf("APN not found")
	}
	return apn, nil
}
