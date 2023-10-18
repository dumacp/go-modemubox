package main

import (
	"fmt"
	"os/exec"
	"time"
)

func ping(ipstest []string, count int, wait, waittime time.Duration) error {

	waitseconds := int(wait.Seconds())
	if waitseconds == 0 {
		waitseconds = 1
	}

	waittimeseconds := int(waittime.Seconds())
	if waittimeseconds == 0 {
		waittimeseconds = 1
	}

	var out []byte
	var err error
	for _, iptest := range ipstest {
		pingcmd := fmt.Sprintf("ping -c %d, -i %d -W %d %s", count, waitseconds, waittimeseconds, iptest)
		out, err = exec.Command("/bin/sh", "-c", pingcmd).Output()
		if err == nil {
			fmt.Println("pint test to ", iptest, "is OK")
			return nil
		}
	}

	return fmt.Errorf("ping cmd error, %q, %w", out, err)

}
