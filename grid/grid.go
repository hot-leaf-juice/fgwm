package grid

import (
	"errors"
	"sync"
	"time"

	"github.com/hot-leaf-juice/fgwm/focus"
	"github.com/hot-leaf-juice/fgwm/view"
	"github.com/hot-leaf-juice/fgwm/wmutils"
)

type Direction int

const (
	Left Direction = iota
	Right
	Up
	Down
)

type FocusStrategy int

const (
	Next FocusStrategy = iota
	Prev
)

type Position struct {
	X, Y int
}

type Size struct {
	W, H int
}

type Rectangle struct {
	TopLeft, BottomRight Position
}

func (p Position) Offset(s Size) Position {
	return Position{p.X + s.W, p.Y + s.H}
}

func (a Position) Diff(b Position) Size {
	return Size{a.X - b.X, a.Y - b.Y}
}

func (a Size) Add(b Size) Size {
	return Size{a.W + b.W, a.H + b.H}
}

func (a Size) Scale(k int) Size {
	return Size{k * a.W, k * a.H}
}

func (r Rectangle) Size() Size {
	return r.BottomRight.Diff(r.TopLeft)
}

func (r Rectangle) Offset(s Size) Rectangle {
	return Rectangle{r.TopLeft.Offset(s), r.BottomRight.Offset(s)}
}

func (r Rectangle) Grow(s Size) Rectangle {
	return Rectangle{r.TopLeft.Offset(s.Scale(-1)), r.BottomRight.Offset(s)}
}

func (r Rectangle) Valid() bool {
	return r.TopLeft.X < r.BottomRight.X && r.TopLeft.Y < r.BottomRight.Y
}

type Grid struct {
	// size of the screen
	screen wmutils.Size
	// margin at edge of screen
	margin wmutils.Size
	// padding around cells
	pad wmutils.Size
	// border around cells
	border wmutils.Pixels
	// size of each cell, including pad and border but excluding margin
	cell wmutils.Size
	// the pixel locations of the cell boundaries
	points map[Position]wmutils.Position
	// the size of the grid in cells
	size  Size
	focus focus.Focus
	view  view.View
	mux   sync.Mutex
}

// The sizes that define the grid layout are made up as follows (bd is border).
// The Y direction is similar.
//
// | <-------------------------------- screen -------------------------------> |
// |        |     |    |      |    |     |     |    |      |    |     |        |
// | margin | pad | bd |      | bd | pad | pad | bd |      | bd | pad | margin |
// |        |     |    |      |    |     |     |    |      |    |     |        |
// |        | <--------- cell ---------> | <--------- cell ---------> |        |
//

type Options struct {
	Border                           wmutils.Pixels
	MinMargin, Pad                   wmutils.Size
	Size                             Size
	InitialView                      int
	FocusTimeout                     time.Duration
	FocussedColour, UnfocussedColour wmutils.Colour
}

func New(opts *Options) (*Grid, error) {
	wid, err := wmutils.Root()
	if err != nil {
		return nil, err
	}
	_, screen, err := wmutils.GetAttributes(wid)
	if err != nil {
		return nil, err
	}
	cell := wmutils.Size{
		W: (screen.W - 2*opts.MinMargin.W) / wmutils.Pixels(opts.Size.W),
		H: (screen.H - 2*opts.MinMargin.H) / wmutils.Pixels(opts.Size.H),
	}
	margin := wmutils.Size{
		W: (screen.W - wmutils.Pixels(opts.Size.W)*cell.W) / 2,
		H: (screen.H - wmutils.Pixels(opts.Size.H)*cell.H) / 2,
	}
	focus, err := focus.New(
		opts.FocusTimeout,
		opts.FocussedColour,
		opts.UnfocussedColour,
	)
	if err != nil {
		return nil, err
	}
	view, err := view.New(screen, opts.Border, opts.InitialView)
	if err != nil {
		return nil, err
	}
	return &Grid{
		screen: screen,
		margin: margin,
		border: opts.Border,
		pad:    opts.Pad,
		cell:   cell,
		size:   opts.Size,
		focus:  focus,
		view:   view,
	}, nil
}

func (g *Grid) WatchWindowEvents() error {
	for ev := range wmutils.WatchEvents() {
		switch ev.Type {
		case wmutils.CreateNotifyEvent:
			// Wait for a tick so that the window's self imposed size has a
			// chance to settle
			time.Sleep(100 * time.Millisecond)
			g.mux.Lock()
			if err := g.centerWID(ev.WID); err != nil {
				return err
			}
			if err := wmutils.SetBorderWidth(ev.WID, g.border); err != nil {
				return err
			}
			g.mux.Unlock()
		case wmutils.DestroyNotifyEvent:
			g.mux.Lock()
			if err := g.focus.Unregister(ev.WID); err != nil {
				return err
			}
			g.view.UnregisterAll(ev.WID)
			g.mux.Unlock()
		case wmutils.UnmapNotifyEvent:
			g.mux.Lock()
			if err := g.focus.Unset(ev.WID); err != nil {
				return err
			}
			g.mux.Unlock()
		case wmutils.MapNotifyEvent:
			g.mux.Lock()
			if err := g.focus.Register(ev.WID); err != nil {
				return err
			}
			g.view.Register(ev.WID)
			g.mux.Unlock()
		}
	}
	return errors.New("Window event channel closed!")
}
