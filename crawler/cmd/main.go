package main

import (
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"

	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const (
	DEFAULT_CITY           = "shanghai"
	DEFAULT_FILE_DAILY     = "../data/{city}-daily"
	DEFAULT_FILE_RESIDENTS = "../data/{city}-residents"
	DEFAULT_FILE_LOG       = "../data/crawler.log"
)

func main() {
	app := &cli.App{
		Name:  "crawler",
		Usage: "用于抓取新冠疫情数据的爬虫",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "pprof",
				Aliases: []string{"p"},
			},
			&cli.BoolFlag{
				Name:    "logfile",
				Aliases: []string{"l"},
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Value:   false,
			},
			&cli.StringFlag{
				Name:    "key_amap",
				EnvVars: []string{"KEY_AMAP"},
			},
			&cli.StringFlag{
				Name:    "key_tianditu",
				EnvVars: []string{"KEY_TIANDITU"},
			},
			&cli.StringFlag{
				Name:    "key_baidu_map",
				EnvVars: []string{"KEY_BAIDU_MAP"},
			},
			&cli.StringFlag{
				Name:  "web_cache",
				Value: "../data/.web_cache",
			},
			&cli.StringFlag{
				Name:  "geo_cache",
				Value: "../data/.geo_cache",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "daily",
				Usage: "抓取每日统计信息",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "city",
						Aliases: []string{"c"},
						Value:   DEFAULT_CITY,
						// Required:    true,
					},
					&cli.BoolFlag{
						Name:  "no-cache",
						Value: false,
					},
					&cli.StringFlag{
						Name:    "daily",
						Aliases: []string{"d"},
						Value:   DEFAULT_FILE_DAILY,
						// Required:    true,
					},
					&cli.StringFlag{
						Name:    "residents",
						Aliases: []string{"r"},
						Value:   DEFAULT_FILE_RESIDENTS,
						// Required:    true,
					},
				},
				Action: actionCrawlDaily,
			},
		},
		Before: func(c *cli.Context) error {
			//	profile
			if c.Bool("pprof") {
				f, err := os.Create("cpu.prof")
				if err != nil {
					log.Fatal("could not create CPU profile: ", err)
				}
				defer f.Close() // error handling omitted for example
				if err := pprof.StartCPUProfile(f); err != nil {
					log.Fatal("could not start CPU profile: ", err)
				}
				defer pprof.StopCPUProfile()

				//	start profile listener
				go func() {
					log.Println(http.ListenAndServe("localhost:3999", nil))
				}()
			}
			//	log to file or stderr
			if c.Bool("logfile") {
				if file, err := os.OpenFile(DEFAULT_FILE_LOG, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666); err != nil {
					log.Errorf("Failed to log to file. %s", err.Error())
				} else {
					log.SetOutput(io.MultiWriter(os.Stderr, file))
				}
			} else {
				log.SetOutput(os.Stderr)
			}
			//	log level
			if c.Bool("verbose") {
				log.SetLevel(log.TraceLevel)
			} else {
				log.SetLevel(log.DebugLevel)
			}
			// log.SetFormatter(&log.TextFormatter{PadLevelText: true, ForceColors: true, FullTimestamp: false})
			return nil
		},
		After: func(c *cli.Context) error {
			if c.Bool("pprof") {
				f, err := os.Create("mem.prof")
				if err != nil {
					log.Fatal("could not create memory profile: ", err)
				}
				defer f.Close() // error handling omitted for example
				runtime.GC()    // get up-to-date statistics
				if err := pprof.WriteHeapProfile(f); err != nil {
					log.Fatal("could not write memory profile: ", err)
				}
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
