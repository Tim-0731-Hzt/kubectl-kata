package plugin

import (
	"context"
	"github.com/Tim-0731-Hzt/kubectl-kata/pkg/kube"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

type DeleteService struct {
	kubeService *kube.KubernetesApiServiceImpl
}

func NewDeleteService() *DeleteService {
	return &DeleteService{}
}
func (d *DeleteService) Complete(cmd *cobra.Command, args []string) error {
	var err error
	d.kubeService, err = kube.NewKubernetesApiServiceImpl()
	if err != nil {
		return err
	}
	return nil
}

func (d *DeleteService) Run() error {
	ctx := context.Background()
	log.Infof("delete kata-deploy")
	if err := d.kubeService.DeleteDaemonSet(ctx, "kata-deploy", "kube-system"); err != nil {
		return err
	}
	cmd := exec.Command("kubectl", "-n", "kube-system", "wait", "--timeout=10m", "--for=delete", "-l", "name=kata-deploy", "pod")
	if err := cmd.Run(); err != nil {
		log.WithError(err).Errorf("failed to execute kubectl wait")
		return err
	}

	log.Infof("create kubelet-kata-cleanup")
	if err := d.kubeService.DeployDaemonSet(ctx, kube.DaemonSetCleanDeployment); err != nil {
		return err
	}

	cmd = exec.Command("kubectl", "-n", "kube-system", "wait", "--timeout=10m", "--for=condition=Ready", "-l", "name=kubelet-kata-cleanup", "pod")
	if err := cmd.Run(); err != nil {
		log.WithError(err).Errorf("failed to execute kubectl wait")
		log.Errorf("delete kubelet-kata-cleanup")
		_ = d.kubeService.DeleteDaemonSet(ctx, "kubelet-kata-cleanup", "kube-system")
		return err
	}

	log.Infof("exec cleanup")
	pods, err := d.kubeService.GetKataDeployPods("name=kubelet-kata-cleanup", "kube-system")
	if err != nil {
		return err
	}
	for _, pod := range pods {
		executeCleanupRequest := kube.ExecCommandRequest{
			PodName:   pod.Name,
			Namespace: pod.Namespace,
			Container: "kube-kata",
			Command:   []string{"bash", "-c", "/opt/kata-artifacts/scripts/kata-deploy.sh reset"},
			StdOut:    os.Stdout,
		}
		if _, err := d.kubeService.ExecuteCommand(executeCleanupRequest); err != nil {
			log.WithError(err).Errorf("failed to execute reset command")
			return err
		}
	}

	log.Infof("delete kubelet-kata-cleanup")
	if err := d.kubeService.DeleteDaemonSet(ctx, "kubelet-kata-cleanup", "kube-system"); err != nil {
		return err
	}
	cmd = exec.Command("kubectl", "-n", "kube-system", "wait", "--timeout=10m", "--for=delete", "-l", "name=kubelet-kata-cleanup", "pod")
	if err := cmd.Run(); err != nil {
		log.WithError(err).Errorf("failed to execute kubectl wait")
		return err
	}
	log.Infof("delete kata-rbac")
	if err := d.kubeService.DeleteRbac(); err != nil {
		return err
	}

	log.Infof("delete kata-runtimeclass")
	for _, s := range []string{"kata-qemu", "kata-clh", "kata-fc", "kata-dragonball"} {
		if err := d.kubeService.DeleteRuntimeClass(s); err != nil {
			return err
		}
	}
	return nil
}
