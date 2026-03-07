package main

import (
	"fmt"
	"log"

	"github.com/MemestaVedas/gobuild/internal/platform/windows"
)

func main() {
	plat := windows.New()
	procs, err := plat.ScanBuildProcesses()
	if err != nil {
		log.Fatalf("Error scanning processes: %v", err)
	}

	fmt.Printf("Total build processes found: %d\n", len(procs))
	for _, p := range procs {
		fmt.Printf("- PID: %d, Name: %s, Cmd: %s\n", p.PID, p.Name, p.CmdLine)
	}
}
