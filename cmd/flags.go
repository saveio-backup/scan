package cmd

import (
	"strings"

	"github.com/oniio/oniDNS/config"
	"github.com/urfave/cli"
)

var (
	//commmon
	LogStderrFlag = cli.BoolFlag{
		Name:  "logstderr",
		Usage: "log to standard error instead of files,default false",
	}
	LogLevelFlag = cli.UintFlag{
		Name:  "loglevel",
		Usage: "Set the log level to `<level>` (0~4). 0:DEBUG 1:INFO 2:WARNING 3:ERROR 4:FATAL",
		Value: config.DEFAULT_LOG_LEVEL,
	}
)

//GetFlagName deal with short flag, and return the flag name whether flag name have short name
func GetFlagName(flag cli.Flag) string {
	name := flag.GetName()
	if name == "" {
		return ""
	}
	return strings.TrimSpace(strings.Split(name, ",")[0])
}
