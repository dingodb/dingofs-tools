package common

import (
	"fmt"

	"github.com/dingodb/dingocli/cli/cli"
	"github.com/dingodb/dingocli/internal/configure/topology"
	"github.com/dingodb/dingocli/internal/task/step"
	"github.com/dingodb/dingocli/internal/task/task"
	tui "github.com/dingodb/dingocli/internal/tui/common"
)

func NewCheckStoreHealthTask(dingocli *cli.DingoCli, dc *topology.DeployConfig) (*task.Task, error) {
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
	t := task.NewTask("Check dingo-store Service", subname, hc.GetSSHConfig())

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

	// check cooridinator leader selection success
	t.AddStep(&step.ContainerExec{
		ContainerId: &containerId,
		Command:     fmt.Sprintf("bash %s/%s", dc.GetProjectLayout().DingoStoreScriptDir, topology.SCRIPT_CHECK_STORE_HEALTH),
		Success:     &success,
		Out:         &out,
		ExecOptions: dingocli.ExecOptions(),
	})

	return t, nil
}
