package common

import (
	"fmt"

	"github.com/dingodb/dingocli/cli/cli"
	"github.com/dingodb/dingocli/internal/configure/topology"
	"github.com/dingodb/dingocli/internal/task/step"
	"github.com/dingodb/dingocli/internal/task/task"
	tui "github.com/dingodb/dingocli/internal/tui/common"
)

func NewSyncJavaOptsTask(dingocli *cli.DingoCli, dc *topology.DeployConfig) (*task.Task, error) {
	serviceId := dingocli.GetServiceId(dc.GetId())
	containerId, err := dingocli.GetContainerId(serviceId)
	if dingocli.IsSkip(dc) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	hc, err := dingocli.GetHost(dc.GetHost())
	if err != nil {
		return nil, err
	}

	// new task
	subname := fmt.Sprintf("host=%s role=%s containerId=%s",
		dc.GetHost(), dc.GetRole(), tui.TrimContainerId(containerId))
	t := task.NewTask("Sync dingo-executor Java Opts", subname, hc.GetSSHConfig())

	// add step to task
	var out string
	var success bool
	host, role := dc.GetHost(), dc.GetRole()
	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      `"{{.ID}}"`,
		Filter:      fmt.Sprintf("id=%s", containerId),
		Out:         &out,
		ExecOptions: dingocli.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: CheckContainerExist(host, role, containerId, &out),
	})

	//t.AddStep(&step.StartContainer{
	//	ContainerId: &containerId,
	//	ExecOptions: dingocli.ExecOptions(),
	//})
	//t.AddStep(&step.Lambda{
	//	Lambda: WaitContainerStart(3),
	//})

	// sync java opts
	t.AddStep(&step.ContainerExec{
		ContainerId: &containerId,
		Command:     fmt.Sprintf("bash %s/%s %s/%s", dc.GetProjectLayout().DingoExecutorBinDir, topology.SCRIPT_SYNC_JAVA_OPTS, dc.GetProjectLayout().DingoExecutorBinDir, topology.SCRIPT_START_EXECUTOR),
		Success:     &success,
		Out:         &out,
		ExecOptions: dingocli.ExecOptions(),
	})

	return t, nil
}
