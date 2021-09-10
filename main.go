package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/lxn/walk"
	wd "github.com/lxn/walk/declarative"
	cron "github.com/robfig/cron/v3"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Application struct {
		Filename     string `yaml:"filename"`
		Fullscreen   string `yaml:"fullscreen"`
		RelativePath string `yaml:"relativepath"`
		Refreshrate  int    `yaml:"refreshrate"`
	} `yaml:"application"`
}

type MyMainWindow struct {
	*walk.MainWindow
	imageView *walk.ImageView
	label     *walk.Label
	prevImage string
}

func checkError(e error) {
	if e != nil {
		log.Fatalf("Error: %v", e)
	}
}

func createFilepath(fp string) {
	if _, err := os.Stat(fp); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(fp, os.ModePerm)
		checkError(err)
	}
}

func removeFileExtension(file []byte) string {
	extension := filepath.Ext(string(file))
	name := string(file)[0 : len(string(file))-len(extension)]
	return name
}

func setupConfig() Config {
	pwd, _ := os.Getwd()
	cfg := Config{}
	configfile, err := ioutil.ReadFile(filepath.Join(pwd, "/config/config.yml"))
	if err != nil {
		defaultConfig := Config{}
		defaultConfig.Application.Filename = "default.txt"
		defaultConfig.Application.Fullscreen = "false"
		defaultConfig.Application.Refreshrate = 10
		defaultConfig.Application.RelativePath = "/resources/"
		data, _ := yaml.Marshal(defaultConfig)
		cfgfp := filepath.Join(pwd, "/config/")
		fp := filepath.Join(pwd, defaultConfig.Application.RelativePath)
		createFilepath(cfgfp)
		createFilepath(fp)
		_ = ioutil.WriteFile(filepath.Join(pwd, "/config/config.yml"), []byte(data), 0644)
		_ = ioutil.WriteFile(filepath.Join(fp, defaultConfig.Application.Filename), []byte("default.png"), 0644)
		cfg = defaultConfig
	} else {
		err = yaml.Unmarshal([]byte(configfile), &cfg)
		checkError(err)
	}
	cfg.Application.RelativePath = filepath.Join(pwd, cfg.Application.RelativePath)
	return cfg
}

func main() {
	var db *walk.DataBinder
	mw := new(MyMainWindow)

	// Destructure configuration
	cfg := setupConfig()
	app := cfg.Application
	Filename, Fullscreen, RelativePath, _ := app.Filename, app.Fullscreen, app.RelativePath, app.Refreshrate
	File := filepath.Join(RelativePath, Filename)

	// Instantiate image
	imgname, _ := ioutil.ReadFile(File)
	imgpth := filepath.Join(RelativePath, string(imgname))
	img, err := walk.NewImageFromFileForDPI(imgpth, 96)
	if err != nil {
		img = nil
		imgname = []byte("No image found")
	}

	// Refresh image
	c := cron.New(cron.WithSeconds())
	c.AddFunc("*/10 * * * * *", func() {
		// runtime.GC()
		// debug.FreeOSMemory()
		imgname, _ := ioutil.ReadFile(File)
		imgpath := filepath.Join(RelativePath, string(imgname))
		if imgpath != mw.prevImage {
			if mw.imageView.Image() != nil {
				mw.imageView.Image().Dispose()
			}
			mw.prevImage = imgpath
			img, err := walk.NewImageFromFileForDPI(imgpath, 96)
			if err != nil {
				mw.imageView.SetImage(nil)
				mw.label.SetText("No image found")
			} else {
				mw.imageView.SetImage(img)
				mw.label.SetText(removeFileExtension(imgname))
			}
		}
	})
	c.Start()

	// Window initialization
	if err := (wd.MainWindow{
		AssignTo: &mw.MainWindow,
		Children: []wd.Widget{
			wd.ImageView{
				AssignTo: &mw.imageView,
				Mode:     wd.ImageViewModeCenter,
			},
			wd.Label{
				AssignTo: &mw.label,
				Font:     wd.Font{Bold: false, Family: "Arial", PointSize: 16},
			},
		},
		DataBinder: wd.DataBinder{
			AssignTo:   &db,
			DataSource: mw,
		},
		Layout:  wd.Grid{Columns: 1, MarginsZero: true},
		MinSize: wd.Size{Width: 400, Height: 300},
		Size:    wd.Size{Width: 800, Height: 600},
		Title:   "Image Viewer",
	}.Create()); err != nil {
		log.Fatal(err)
	}
	mw.imageView.SetImage(img)
	mw.label.SetText(removeFileExtension(imgname))
	fullscreen, err := strconv.ParseBool(Fullscreen)
	if err != nil {
		fullscreen = false
	}
	mw.SetFullscreen(fullscreen)
	mw.Run()
}
