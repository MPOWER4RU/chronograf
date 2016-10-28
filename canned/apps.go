package canned

//go:generate go-bindata -o apps_gen.go -ignore README|apps -pkg canned .

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/influxdata/chronograf"
	clog "github.com/influxdata/chronograf/log"
	"golang.org/x/net/context"
)

const AppExt = ".json"

var logger = clog.New()

// Apps are canned JSON layouts.  Implements LayoutStore.
type Apps struct {
	Dir      string                                      // Dir is the directory contained the pre-canned applications.
	Load     func(string) (chronograf.Layout, error)     // Load loads string name and return a Layout
	Filename func(string, chronograf.Layout) string      // Filename takes dir and layout and returns loadable file
	Create   func(string, chronograf.Layout) error       // Create will write layout to file.
	ReadDir  func(dirname string) ([]os.FileInfo, error) // ReadDir reads the directory named by dirname and returns a list of directory entries sorted by filename.
	Remove   func(name string) error                     // Remove file
	IDs      chronograf.ID                               // IDs generate unique ids for new application layouts
}

func NewApps(dir string, ids chronograf.ID) chronograf.LayoutStore {
	return &Apps{
		Dir:      dir,
		Load:     loadFile,
		Filename: fileName,
		Create:   createLayout,
		ReadDir:  ioutil.ReadDir,
		Remove:   os.Remove,
		IDs:      ids,
	}
}

func fileName(dir string, layout chronograf.Layout) string {
	base := fmt.Sprintf("%s_%s%s", layout.Measurement, layout.ID, AppExt)
	return path.Join(dir, base)
}

func loadFile(name string) (chronograf.Layout, error) {
	octets, err := ioutil.ReadFile(name)
	if err != nil {
		logger.
			WithField("component", "apps").
			WithField("name", name).
			Error("Unable to read file")
		return chronograf.Layout{}, err
	}
	var layout chronograf.Layout
	if err = json.Unmarshal(octets, &layout); err != nil {
		logger.
			WithField("component", "apps").
			WithField("name", name).
			Error("File is not a layout")
		return chronograf.Layout{}, err
	}
	return layout, nil
}

func createLayout(file string, layout chronograf.Layout) error {
	h, err := os.Create(file)
	if err != nil {
		return err
	}
	defer h.Close()
	if octets, err := json.MarshalIndent(layout, "    ", "    "); err != nil {
		logger.
			WithField("component", "apps").
			WithField("name", file).
			Error("Unable to marshal layout:", err)
		return err
	} else {
		if _, err := h.Write(octets); err != nil {
			logger.
				WithField("component", "apps").
				WithField("name", file).
				Error("Unable to write layout:", err)
			return err
		}
	}
	return nil
}

func (a *Apps) All(ctx context.Context) ([]chronograf.Layout, error) {
	files, err := a.ReadDir(a.Dir)
	if err != nil {
		return nil, err
	}

	layouts := []chronograf.Layout{}
	for _, file := range files {
		if path.Ext(file.Name()) != AppExt {
			continue
		}
		if layout, err := a.Load(path.Join(a.Dir, file.Name())); err != nil {
			continue // We want to load all files we can.
		} else {
			layouts = append(layouts, layout)
		}
	}
	return layouts, nil
}

func (a *Apps) Add(ctx context.Context, layout chronograf.Layout) (chronograf.Layout, error) {
	var err error
	layout.ID, err = a.IDs.Generate()
	file := a.Filename(a.Dir, layout)
	if err = a.Create(file, layout); err != nil {
		return chronograf.Layout{}, err
	}
	return layout, nil
}

func (a *Apps) Delete(ctx context.Context, layout chronograf.Layout) error {
	file, err := a.idToFile(layout.ID)
	if err != nil {
		return err
	}

	if err := a.Remove(file); err != nil {
		logger.
			WithField("component", "apps").
			WithField("name", file).
			Error("Unable to remove layout:", err)
		return err
	}
	return nil
}

func (a *Apps) Get(ctx context.Context, ID string) (chronograf.Layout, error) {
	file, err := a.idToFile(ID)
	if err != nil {
		return chronograf.Layout{}, err
	}
	l, err := a.Load(file)
	if err != nil {
		return chronograf.Layout{}, chronograf.ErrLayoutNotFound
	}
	return l, nil
}

func (a *Apps) Update(ctx context.Context, layout chronograf.Layout) error {
	file, err := a.idToFile(layout.ID)
	if err != nil {
		return err
	}

	l, err := a.Load(file)
	if err != nil {
		return chronograf.ErrLayoutNotFound
	}

	if err := a.Delete(ctx, l); err != nil {
		return err
	}
	file = a.Filename(a.Dir, layout)
	return a.Create(file, layout)
}

// idToFile takes an id and finds the associated filename
func (a *Apps) idToFile(ID string) (string, error) {
	// Because the entire layout information is not known at this point, we need
	// to try to find the name of the file through matching.
	files, err := a.ReadDir(a.Dir)
	if err != nil {
		return "", err
	}

	var file string
	for _, f := range files {
		if strings.Contains(f.Name(), ID) {
			file = path.Join(a.Dir, f.Name())
		}
	}
	if file == "" {
		return "", chronograf.ErrLayoutNotFound
	}
	return file, nil
}
