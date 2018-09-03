package oyster_utils

import (
	"fmt"
	"github.com/stackimpact/stackimpact-go"
	"os"
	"runtime"
)

var Agent *stackimpact.Agent

func init() {
	Agent = stackimpact.Start(stackimpact.Options{
		AgentKey: "fa6aa0f39c917c329749721e98cfb6b269b88ed9",
		AppName:  "Brokernode",
		HostName: os.Getenv("HOST_IP"),
	})
}

func Trace() string {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	return fmt.Sprintf("%s\n", f.Name())
}
