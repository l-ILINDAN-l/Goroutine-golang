package main

import (
	"fmt"
	"github.com/beevik/ntp"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New().WithFields(logrus.Fields{"module": "L2", "function": "main"})

	time, err := ntp.Time("0.beevik-ntp.pool.ntp.org")
	if err != nil {
		logger.Fatalf("error get time: %v", err)
	}
	fmt.Println(time)
}
