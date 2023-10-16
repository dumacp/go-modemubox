package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dumacp/go-modemubox"
)

func main() {

	ch, err := modemubox.TailKmesg(context.TODO())
	if err != nil {
		log.Fatalln(err)
	}

	for v := range ch {
		fmt.Println("line: ", v)
	}

}
