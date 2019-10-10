package main

import (
	"log"
	"os"

	"github.com/nocquidant/mageproj/mgl"
	"github.com/nocquidant/mageproj/mgp"
)

func workdir() string {
	workdir, err := os.Getwd()
	if err != nil {
		workdir = "."
	}
	return workdir
}

func main() {
	lib := mgl.NewMageLibrary(workdir())

	log.Printf("Project version is: %s\n", lib.Version())

	proj := &mgp.MageProject{ProjectName: "aReallyGoodOne"}
	proj = mgp.InitMageProject(workdir(), proj)

	log.Printf("Project name is: %s\n", proj.ProjectName)
	proj.PrintInfo()
}
