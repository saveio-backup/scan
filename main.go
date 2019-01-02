package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/oniio/oniDNS/cmd"
	"github.com/oniio/oniDNS/common/log"
	"github.com/oniio/oniDNS/config"
	"github.com/oniio/oniDNS/netserver"
	"github.com/urfave/cli"
)

func initAPP() *cli.App {
	app := cli.NewApp()
	app.Usage = "dasein seeds"
	app.Action = seed
	app.Version = config.VERSION
	app.Copyright = "Copyright in 2018 The Dasein Authors"
	app.Commands = []cli.Command{}
	app.Flags = []cli.Flag{
		//common setting
		cmd.LogStderrFlag,
		cmd.LogLevelFlag,
	}
	app.Before = func(context *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		return nil
	}
	return app
}

func main() {
	if err := initAPP().Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func seed(ctx *cli.Context) {
	initLog(ctx)

	err := initConfig(ctx)
	if err != nil {
		log.Errorf("initConfig error:%s", err)
		return
	}
	svr := netserver.NewNetServer()
	// svr := &dsp.Server{
	// 	State: &dsp.Info{},
	// }
	if err = svr.Run(); err != nil {
		log.Errorf("run ddns server error:%s", err)
		return
	}

	waitToExit()
}

func initConfig(ctx *cli.Context) error {
	_, err := cmd.SetSeedsConfig(ctx)
	if err != nil {
		return err
	}
	log.Infof("Config init success")
	return nil
}

func initLog(ctx *cli.Context) {
	//init log module
	log.SetLevel(ctx.GlobalUint(cmd.GetFlagName(cmd.LogLevelFlag)))
	log.SetMaxSize(config.DEFAULT_MAX_LOG_SIZE)
	if ctx.Bool(cmd.GetFlagName(cmd.LogStderrFlag)) {
		log.InitLog(0, config.DEFAULT_LOG_DIR)
	} else {
		log.InitLog(1, config.DEFAULT_LOG_DIR)
	}
	log.Info("start logging...")
}

func waitToExit() {
	exit := make(chan bool, 0)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for sig := range sc {
			log.Infof("seeds received exit signal:%v.", sig.String())
			close(exit)
			break
		}
	}()
	<-exit
}
