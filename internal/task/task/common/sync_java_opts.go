package common

import (
	"fmt"

	"github.com/dingodb/dingofs-tools/cli/cli"
	"github.com/dingodb/dingofs-tools/internal/configure/topology"
	"github.com/dingodb/dingofs-tools/internal/task/step"
	"github.com/dingodb/dingofs-tools/internal/task/task"
	tui "github.com/dingodb/dingofs-tools/internal/tui/common"
)

func NewSyncJavaOptsTask(dingoadm *cli.DingoAdm, dc *topology.DeployConfig) (*task.Task, error) {
	serviceId := dingoadm.GetServiceId(dc.GetId())
	containerId, err := dingoadm.GetContainerId(serviceId)
	if dingoadm.IsSkip(dc) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	hc, err := dingoadm.GetHost(dc.GetHost())
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
		ExecOptions: dingoadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: CheckContainerExist(host, role, containerId, &out),
	})

	//t.AddStep(&step.StartContainer{
	//	ContainerId: &containerId,
	//	ExecOptions: dingoadm.ExecOptions(),
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
		ExecOptions: dingoadm.ExecOptions(),
	})

	return t, nil
}
