package main

import (
	"flag"
	"fmt"
	pluginpb "github.com/dsrvlabs/vatz-proto/plugin/v1"
	"github.com/dsrvlabs/vatz/sdk"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/cpu"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/structpb"
	"os"
	"time"
)

const (
	// Default values.
	defaultAddr = "127.0.0.1"
	defaultPort = 9091
	pluginName  = "machine-status-cpu"
	methodName  = "GetMachineCPUUsage"
)

var (
	addr string
	port int
)

func init() {
	flag.StringVar(&addr, "addr", defaultAddr, "IP Address(e.g. 0.0.0.0, 127.0.0.1)")
	flag.IntVar(&port, "port", defaultPort, "Port number, default 9091")
	flag.Parse()
}

func main() {
	p := sdk.NewPlugin(pluginName)
	p.Register(pluginFeature)
	ctx := context.Background()
	if err := p.Start(ctx, addr, port); err != nil {
		fmt.Println("exit")
	}
}

func pluginFeature(info, option map[string]*structpb.Value) (sdk.CallResponse, error) {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	state := pluginpb.STATE_SUCCESS
	severity := pluginpb.SEVERITY_INFO

	_, err := cpu.Info()

	if err != nil {
		state = pluginpb.STATE_FAILURE
		severity = pluginpb.SEVERITY_ERROR
		log.Error().
			Str(methodName, "Getting CPU info has failed.").
			Msg(pluginName)

	}

	totalUsed := 0.0
	percent, _ := cpu.Percent(3*time.Second, false)

	for _, numb := range percent {
		totalUsed += numb
	}

	cpuScale := 0
	if totalUsed < 50.0 {
		cpuScale = 1
	} else if totalUsed < 65 {
		cpuScale = 2
	} else if totalUsed < 90 {
		cpuScale = 3
	} else {
		cpuScale = 4
	}

	if state == pluginpb.STATE_SUCCESS {
		if cpuScale > 3 {
			severity = pluginpb.SEVERITY_CRITICAL
		} else if cpuScale > 2 {
			severity = pluginpb.SEVERITY_WARNING
		}
	}
	contentMSG := "Total CPU Usage: " + fmt.Sprintf("%.2f", totalUsed) + "%"

	log.Info().
		Str(methodName, contentMSG).
		Msg(pluginName)

	ret := sdk.CallResponse{
		FuncName:   methodName,
		Message:    contentMSG,
		Severity:   severity,
		State:      state,
		AlertTypes: []pluginpb.ALERT_TYPE{pluginpb.ALERT_TYPE_DISCORD},
	}

	return ret, nil
}
