package modemubox

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TailKmesg(ctx context.Context) (chan string, error) {

	f, err := os.Open("/dev/kmsg")
	if err != nil {
		return nil, err
	}

	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return nil, err
	}

	watcher := time.NewTicker(1 * time.Second)

	reader := bufio.NewReader(f)

	ch := make(chan string)

	go func() {
		defer f.Close()
		defer watcher.Stop()

		for {
			select {
			case <-ctx.Done():
				fmt.Println("close kmesg read")
				return
			case <-watcher.C:
				line, err := reader.ReadString('\n')
				if err != nil {
					fmt.Printf("error reading line: %s", err)
					return
				}
				fmt.Println("line form kmesg ", line) // process the line
				select {
				case <-ctx.Done():
				case ch <- line:
				default:
				}
			}
		}
	}()

	return ch, nil
}

func TailKmesg_DEPRECATED(ctx context.Context) (chan string, error) {

	f, err := os.Open("/dev/kmsg")
	if err != nil {
		return nil, err
	}

	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("error creating watcher: %w", err)
	}

	err = watcher.Add("/dev/kmsg")
	if err != nil {
		return nil, fmt.Errorf("error adding /dev/kmsg to watcher: %w", err)
	}

	reader := bufio.NewReader(f)

	ch := make(chan string)

	go func() {
		defer f.Close()
		defer watcher.Close()

		for {
			select {
			case <-ctx.Done():
				fmt.Println("close kmesg read")
				return
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					line, err := reader.ReadString('\n')
					if err != nil {
						fmt.Printf("error reading line: %s", err)
						return
					}
					fmt.Println("line form kmesg ", line) // process the line
					select {
					case <-ctx.Done():
					case ch <- line:
					default:
					}
				}
			case err := <-watcher.Errors:
				fmt.Printf("watcher error: %s\n", err)
				return
			}
		}
	}()

	return ch, nil
}
