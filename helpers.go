package main

import (
	"fmt"
	"strconv"

	"github.com/tideland/golib/logger"
)

func toFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	exitIfErr(err, fmt.Sprintf("Could not convert %s to float: %s", s))

	return f
}

func toInt(s, errorMsg string) int {
	i, err := strconv.ParseInt(s, 10, 64)
	exitIfErr(err, fmt.Sprintf("could not parse %s", errorMsg))
	return int(i)
}

func exitIfErr(err error, message string) {
	if err != nil {
		logger.Fatalf(message)
	}
}
