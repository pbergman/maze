package config

import (
	"flag"
	"strconv"
	"time"
)

type AppConfig struct {
	Width    int
	Height   int
	Scale    uint
	Files    struct{ Raw, Solved, Animation string }
	Server   bool
	Port     int
	Template string
}

var Config *AppConfig

func init() {
	Config = &AppConfig{}
	flag.IntVar(&Config.Height, "h", 10, "Height of maze")
	flag.IntVar(&Config.Width, "w", 10, "Width of maze")
	flag.UintVar(&Config.Scale, "s", 1, "Pixel scale ratio")
	flag.BoolVar(&Config.Server, "S", false, "Start web server")
	flag.IntVar(&Config.Port, "p", 8080, "Port to listen for server")
	flag.StringVar(&Config.Template, "t", "template/base.html", "Template location for server")
	flag.StringVar(&Config.Files.Raw, "rf", "rmaze."+strconv.Itoa(int(time.Now().Unix()))+".gif", "File name to write maze to")
	flag.StringVar(&Config.Files.Solved, "sf", "smaze."+strconv.Itoa(int(time.Now().Unix()))+".gif", "File name to write solved maze to")
	flag.StringVar(&Config.Files.Animation, "af", "amaze."+strconv.Itoa(int(time.Now().Unix()))+".gif", "File name to write anmation maze to")
	flag.Parse()
}