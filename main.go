package main

import (
	"log"

	"github.com/voyages-sncf-technologies/mageproj/mgl"
	"github.com/voyages-sncf-technologies/mageproj/mgp"
)

func main() {
	lib := mgl.NewMageLibrary(".")
	log.Printf("Project version is: %s\n", lib.Version())

	proj := mgp.NewMageProject(".", "agoodone", "github.com/name/proj")
	log.Printf(proj.PrintInfo())
}
