package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/hot-leaf-juice/fgwm/grid"
	"github.com/hot-leaf-juice/fgwm/wmutils"
)

type oneShotClient struct {
	name     string
	command  string
	windowID wmutils.WindowID
	grid     *grid.Grid
}

func main() {
	wid, err := wmutils.Focussed()
	if err != nil {
		log.Fatal(err)
	}
	g, err := grid.New(&grid.Options{
		Border:    wmutils.Size{5, 5},
		MinMargin: wmutils.Size{10, 10},
		Pad:       wmutils.Size{10, 10},
		Size:      grid.Size{24, 24},
	})
	if err != nil {
		log.Fatal(err)
	}
	c := &oneShotClient{name: os.Args[0], windowID: wid, grid: g}
	if len(os.Args) < 2 {
		c.printHelpAndExit(nil)
	}
	c.command = os.Args[1]
	switch c.command {
	case "snap":
		c.noArgsHandler(c.grid.Snap, os.Args[2:])
	case "move":
		c.sizeArgHandler(c.grid.Move, os.Args[2:])
	case "grow":
		c.sizeArgHandler(c.grid.Grow, os.Args[2:])
	case "center":
		c.noArgsHandler(c.grid.Center, os.Args[2:])
	case "throw":
		c.directionArgHandler(c.grid.Throw, os.Args[2:])
	case "spread":
		c.directionArgHandler(c.grid.Spread, os.Args[2:])
	case "teleport":
		c.handleTeleport(os.Args[2:])
	case "help":
		c.printHelpAndExit(os.Args[2:])
	default:
		c.printHelpAndExit(nil)
	}
}

func (c *oneShotClient) noArgsHandler(
	f func(wid wmutils.WindowID) error,
	args []string,
) {
	if len(args) != 0 {
		c.printHelpAndExit(args)
	}
	if err := f(c.windowID); err != nil {
		log.Fatal(err)
	}
}

func (c *oneShotClient) sizeArgHandler(
	f func(wid wmutils.WindowID, size grid.Size) error,
	args []string,
) {
	if len(args) != 2 {
		c.printHelpAndExit(args)
	}
	var size grid.Size
	var err error
	if size.W, err = strconv.Atoi(args[0]); err != nil {
		c.printHelpAndExit(args)
	}
	if size.H, err = strconv.Atoi(args[1]); err != nil {
		c.printHelpAndExit(args)
	}
	if err := f(c.windowID, size); err != nil {
		log.Fatal(err)
	}
}

func (c *oneShotClient) directionArgHandler(
	f func(wid wmutils.WindowID, direction grid.Direction) error,
	args []string,
) {
	if len(args) != 1 {
		c.printHelpAndExit(args)
	}
	var direction grid.Direction
	switch strings.ToLower(args[0]) {
	case "left", "l", "west", "w":
		direction = grid.Left
	case "right", "r", "east", "e":
		direction = grid.Right
	case "up", "u", "north", "n":
		direction = grid.Up
	case "down", "d", "south", "s":
		direction = grid.Down
	default:
		c.printHelpAndExit(args)
	}
	if err := f(c.windowID, direction); err != nil {
		log.Fatal(err)
	}
}

func (c *oneShotClient) handleTeleport(args []string) {
	if len(args) != 4 {
		c.printHelpAndExit(args)
	}
	var tl, br grid.Position
	var err error
	if tl.X, err = strconv.Atoi(args[0]); err != nil {
		c.printHelpAndExit(args)
	}
	if tl.Y, err = strconv.Atoi(args[1]); err != nil {
		c.printHelpAndExit(args)
	}
	if br.X, err = strconv.Atoi(args[2]); err != nil {
		c.printHelpAndExit(args)
	}
	if br.Y, err = strconv.Atoi(args[3]); err != nil {
		c.printHelpAndExit(args)
	}
	if err := c.grid.Teleport(c.windowID, tl, br); err != nil {
		log.Fatal(err)
	}
}

func (c *oneShotClient) printHelpAndExit(args []string) {
	// TODO improve this (including command specific help)
	if len(args) == 0 {
		fmt.Printf("Usage:\n\n\t%v command [arguments]\n\n", c.name)
		fmt.Println("Where command is one of:\n")
		fmt.Println("\tsnap, move, grow, center, throw, spred, teleport, help")
	} else {
		fmt.Printf("help for %v coming soon!\n", args[0])
	}
	os.Exit(0)
}
