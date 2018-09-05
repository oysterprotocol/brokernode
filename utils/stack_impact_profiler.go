package oyster_utils

import (
	"fmt"
	"github.com/stackimpact/stackimpact-go"
	"os"
	"runtime"
)

/*Agent - the stackImpact Agent*/
var Agent *stackimpact.Agent

var agentKey string
var span *stackimpact.Span

func init() {

	agentKey = os.Getenv("STACK_IMPACT_KEY")

	if agentKey != "" {
		Agent = stackimpact.Start(stackimpact.Options{
			AgentKey: agentKey,
			AppName:  "Brokernode",
			HostName: os.Getenv("HOST_IP"),
		})
	}
}

/*StartProfile starts stackImpact profiling*/
func StartProfile() {
	if agentKey != "" {
		span = Agent.Profile()
	}
}

/*StartProfileWithName starts stackImpact profiling with a specific name*/
func StartProfileWithName(name string) {
	if agentKey != "" {
		span = Agent.ProfileWithName(name)
	}
}

/*StopProfile stops stackImpact profiling*/
func StopProfile() {
	if agentKey != "" {
		span.Stop()
	}
}

/*Trace should return the name of the calling function*/
func Trace() string {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	return fmt.Sprintf("%s\n", f.Name())
}
