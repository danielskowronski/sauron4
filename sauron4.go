package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"time"

	"github.com/go-ping/ping"
	client "github.com/influxdata/influxdb1-client"
	"gopkg.in/yaml.v2"
)

// Config is struct of config file
type Config struct {
	InfluxDB struct {
		Enabled bool   `yaml:"enabled"`
		Host    string `yaml:"host"`
		Port    int    `yaml:"port"`
		DB      string `yaml:"db"`
		User    string `yaml:"user"`
		Pass    string `yaml:"pass"`
	} `yaml:"influxdb"`
	WhoAmI       string `yaml:"whoami"`
	PingerParams struct {
		ICMP struct {
			Count    int `yaml:"count"`
			Interval int `yaml:"interval"`
			Timeout  int `yaml:"timeout"`
		} `yaml:"icmp"`
	} `yaml:"pinger_params"`
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
		pinger.Count = config.PingerParams.ICMP.Count
		pinger.Interval = time.Duration(config.PingerParams.ICMP.Interval * 1000000)
		pinger.Timeout = time.Duration(config.PingerParams.ICMP.Timeout * 1000000)
		pingErr = pinger.Run()
		if pingErr != nil {
			panic(pingErr)
		}
		stats := pinger.Statistics()

		fmt.Printf("%s_%s: loss %.2f%%, rtt %dms\n", config.WhoAmI, target.Name, stats.PacketLoss, stats.AvgRtt.Milliseconds())

		if config.InfluxDB.Enabled {
			host, hostParseErr := url.Parse(fmt.Sprintf("http://%s:%d", config.InfluxDB.Host, config.InfluxDB.Port))
			if hostParseErr != nil {
				fmt.Println(hostParseErr)
				continue
			}
			c, connErr := client.NewClient(client.Config{
				URL:      *host,
				Username: config.InfluxDB.User,
				Password: config.InfluxDB.Pass,
			})
			if connErr != nil {
				fmt.Println("Error creating InfluxDB Client: ", connErr.Error())
				continue
			}

			pts := make([]client.Point, 1)
			pts[0] = client.Point{
				Measurement: fmt.Sprintf("%s_%s", config.WhoAmI, target.Name),
				Fields: map[string]interface{}{
					"rtt":  stats.AvgRtt.Milliseconds(),
					"loss": stats.PacketLoss,
				},
				Time:      time.Now(),
				Precision: "ns",
			}

			bps := client.BatchPoints{
				Points:   pts,
				Database: config.InfluxDB.DB,
			}
			_, writeErr := c.Write(bps)
			if writeErr != nil {
				fmt.Println(writeErr)
				continue
			}
		}
	}
}
