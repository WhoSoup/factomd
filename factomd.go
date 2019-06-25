// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/FactomProject/factomd/config"

	"runtime"
	"time"

	. "github.com/FactomProject/factomd/engine"
)

func main() {
	// uncomment StartProfiler() to run the pprof tool (for testing)

	//  Go Optimizations...
	runtime.GOMAXPROCS(runtime.NumCPU()) // TODO: should be *2 to use hyperthreadding? -- clay

	cfg := config.LoadConfig()
	cfg.PrintSettings(true)

	state := Factomd(cfg)
	for state.Running() {
		time.Sleep(time.Second)
	}
	fmt.Println("Waiting to Shut Down")
	time.Sleep(time.Second * 5)
}
