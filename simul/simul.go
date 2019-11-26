package main

import (
	"go.dedis.ch/onet/v3"
	"go.dedis.ch/onet/v3/simul"
)

func main() {
	onet.SimulationRegister("ServiceDrynx", NewSimulationDrynx)
	simul.Start()
}
