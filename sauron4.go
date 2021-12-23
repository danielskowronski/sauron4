package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/go-ping/ping"
	"gopkg.in/yaml.v2"
)

// Config is struct of config file
type Config struct {
	InfluxDB struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
		DB   string `yaml:"db"`
		User string `yaml:"user"`
		Pass string `yaml:"pass"`
	} `yaml:"influxdb"`
	WhoAmI  string `yaml:"whoami"`
	Targets []struct {
		Name string `yaml:"name"`
		Type string `yaml:"type"`
		Host string `yaml:"host"`
	} `yaml:"targets"`
}

func main() {
	yamlFile, readFileErr := ioutil.ReadFile(os.Args[1])
	if readFileErr != nil {
		panic(readFileErr)
	}

	var config Config
	configParseErr := yaml.Unmarshal(yamlFile, &config)
	if configParseErr != nil {
		panic(configParseErr)
	}

	for _, target := range config.Targets {
		pinger, pingErr := ping.NewPinger(target.Host)
		if pingErr != nil {
			panic(pingErr)
		}
		pinger.Count = 3
		pinger.Interval, _ = time.ParseDuration("10ms")
		pinger.Timeout, _ = time.ParseDuration("500ms")
		pingErr = pinger.Run()
		if pingErr != nil {
			panic(pingErr)
		}
		stats := pinger.Statistics()

		fmt.Printf("%s_%s: loss %.2f%%, rtt %dms\n", config.WhoAmI, target.Name, stats.PacketLoss, stats.AvgRtt.Milliseconds())
	}
}
