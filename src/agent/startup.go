package main

import (
	sp "github.com/gonet2/libs/services"
)

func startup() {
	go sig_handler()
	// init services discovery
	sp.Init()
}
