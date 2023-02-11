package plugin

import (
	"fmt"
	"github.com/Tim-0731-Hzt/kubectl-kata/pkg/kube"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

type ConfigService struct {
	kubeService  *kube.KubernetesApiServiceImpl
	DebugConsole bool
}

func NewConfigService() *ConfigService {
	return &ConfigService{}
}

func (c *ConfigService) Complete(cmd *cobra.Command, args []string) error {
	var err error
	c.kubeService, err = kube.NewKubernetesApiServiceImpl()
	if err != nil {
		return err
	}
	return nil
}
func (c *ConfigService) Validate() error {
	cmd := exec.Command("kubectl", "-n", "kube-system", "wait", "--timeout=10m", "--for=condition=Ready", "-l", "name=kata-deploy", "pod")
	if err := cmd.Run(); err != nil {
		log.WithError(err).Errorf("kata deploy not ready")
		return err
	}
	return nil
}
func (c *ConfigService) Run() error {
	var shellScript string
	if c.DebugConsole {
		shellScript = fmt.Sprintf(`
		for var in qemu clh fc dragonball
		do
			sed -i 's/.*debug_console_enabled.*/debug_console_enabled = true/' /opt/kata/share/defaults/kata-containers/configuration-$var.toml
		done
		`,
		)
	} else {
		shellScript = fmt.Sprintf(`
		for var in qemu clh fc dragonball
		do
		   sed -i 's/.*debug_console_enabled.*/#debug_console_enabled = true/' /opt/kata/share/defaults/kata-containers/configuration-$var.toml
		done
		`,
		)
	}
	pods, err := c.kubeService.GetKataDeployPods("name=kata-deploy", "kube-system")
	if err != nil {
		return err
	}
	for _, pod := range pods {
		executeCleanupRequest := kube.ExecCommandRequest{
			PodName:   pod.Name,
			Namespace: pod.Namespace,
			Container: "kube-kata",
			Command:   []string{"/bin/sh", "-c", shellScript},
			StdOut:    os.Stdout,
		}
		if _, err := c.kubeService.ExecuteCommand(executeCleanupRequest); err != nil {
			log.WithError(err).Errorf("failed to config debug console")
			return err
		}
	}
	return nil
}
