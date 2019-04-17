package main

import (
	"fmt"
	"time"

	"github.com/chenlei87/gocommon/rps"
)

func main() {
	nt := rps.Newrps()

	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			nt.Add(10)

		}
	}()

	//读取
	go func() {
		for {
			time.Sleep(1 * time.Second)
			fmt.Println(nt.Average().String())

			fmt.Println(nt.String())

		}
	}()

	stop := make(chan struct{})
	<-stop

}
