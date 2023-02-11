package plugin

import (
	"context"
	"github.com/Tim-0731-Hzt/kubectl-kata/pkg/kube"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os/exec"
)

type DeployService struct {
	kubeService *kube.KubernetesApiServiceImpl
}

func NewDeployService() *DeployService {
	return &DeployService{}
}

func (d *DeployService) Complete(cmd *cobra.Command, args []string) error {
	var err error
	d.kubeService, err = kube.NewKubernetesApiServiceImpl()
	if err != nil {
		return err
	}
	return nil
}

func (d *DeployService) Validate() error {
	return nil
}

func (d *DeployService) Run() error {
	ctx := context.Background()
	log.Infof("create kata-rbac")
	if err := d.kubeService.CreateRbac(ctx, kube.ServiceAccount, kube.ClusterRole, kube.ClusterRoleBinding); err != nil {
		return err
	}
	log.Infof("create kata-deploy")
	if err := d.kubeService.DeployDaemonSet(ctx, kube.DaemonSetDeployment); err != nil {
		return err
	}
	cmd := exec.Command("kubectl", "-n", "kube-system", "wait", "--timeout=10m", "--for=condition=Ready", "-l", "name=kata-deploy", "pod")
	if err := cmd.Run(); err != nil {
		log.WithError(err).Errorf("failed to execute kubectl wait")
		return err
	}
	log.Infof("create kata-runtimeclass")
	QemuRuntimeClass := kube.RuntimeClass("kata-qemu", "250m", "160Mi")
	if err := d.kubeService.CreateRuntimeClass(ctx, QemuRuntimeClass); err != nil {
		return err
	}
	ClhRuntimeClass := kube.RuntimeClass("kata-clh", "250m", "130Mi")
	if err := d.kubeService.CreateRuntimeClass(ctx, ClhRuntimeClass); err != nil {
		return err
	}
	FcRuntimeClass := kube.RuntimeClass("kata-fc", "250m", "130Mi")
	if err := d.kubeService.CreateRuntimeClass(ctx, FcRuntimeClass); err != nil {
		return err
	}
	DragonballRuntimeClass := kube.RuntimeClass("kata-dragonball", "250m", "130Mi")
	if err := d.kubeService.CreateRuntimeClass(ctx, DragonballRuntimeClass); err != nil {
		return err
	}
	log.Infof("ready to go now")
	return nil
}
