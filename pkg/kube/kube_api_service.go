package kube

import (
	"context"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	apps_v1 "k8s.io/api/apps/v1"
	api_v1 "k8s.io/api/core/v1"
	node_v1 "k8s.io/api/node/v1"
	rbac "k8s.io/api/rbac/v1"
	k_error "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	utilexec "k8s.io/client-go/util/exec"
	"k8s.io/kubectl/pkg/scheme"
	"os"
	"time"
)

var KubernetesConfigFlags = genericclioptions.NewConfigFlags(true)

type KubernetesApiServiceImpl struct {
	clientset  *kubernetes.Clientset
	restConfig *rest.Config
}

type ExecCommandRequest struct {
	PodName   string
	Namespace string
	Container string
	Command   []string
	StdIn     io.Reader
	StdOut    io.Writer
	StdErr    io.Writer
}

func NewKubernetesApiServiceImpl() (k *KubernetesApiServiceImpl, err error) {
	k = &KubernetesApiServiceImpl{}
	k.restConfig, err = KubernetesConfigFlags.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	k.restConfig.Timeout = 30 * time.Second
	k.clientset, err = kubernetes.NewForConfig(k.restConfig)
	if err != nil {
		return nil, err
	}
	return k, nil
}

func (k *KubernetesApiServiceImpl) GetPod(ctx context.Context, podName string, namespace string) (*api_v1.Pod, error) {
	return k.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
}

func (k *KubernetesApiServiceImpl) GetKataDeployPod(p *api_v1.Pod) (*api_v1.Pod, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: "name=kata-deploy",
	}
	pods, err := k.clientset.CoreV1().Pods("kube-system").List(context.TODO(), listOptions)
	if err != nil {
		return nil, err
	}
	for _, pod := range pods.Items {
		if pod.Spec.NodeName == p.Spec.NodeName {
			return &pod, nil
		}
	}
	return nil, err
}

func (k *KubernetesApiServiceImpl) GetKataDeployPods(labelSelector string, namespace string) ([]api_v1.Pod, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}
	pods, err := k.clientset.CoreV1().Pods(namespace).List(context.TODO(), listOptions)
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

func (k *KubernetesApiServiceImpl) ExecuteCommand(req ExecCommandRequest) (int, error) {
	execRequest := k.clientset.CoreV1().RESTClient().Post().Resource("pods").Name(req.PodName).Namespace(req.Namespace).SubResource("exec")
	execRequest.VersionedParams(&api_v1.PodExecOptions{
		Container: req.Container,
		Command:   req.Command,
		Stdin:     req.StdIn != nil,
		Stdout:    req.StdOut != nil,
		TTY:       false,
	}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(k.restConfig, "POST", execRequest.URL())
	if err != nil {
		return 0, nil
	}
	err = exec.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdout: req.StdOut,
		Tty:    false,
	})
	var exitCode = 0
	if err != nil {
		if exitErr, ok := err.(utilexec.ExitError); ok && exitErr.Exited() {
			exitCode = exitErr.ExitStatus()
			return 1, err
		}
	}
	return exitCode, nil
}

func (k *KubernetesApiServiceImpl) ExecuteVMCommand(req ExecCommandRequest) (int, error) {
	execRequest := k.clientset.CoreV1().RESTClient().Post().Resource("pods").Name(req.PodName).Namespace(req.Namespace).SubResource("exec")
	execRequest.VersionedParams(&api_v1.PodExecOptions{
		Container: req.Container,
		Command:   req.Command,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(k.restConfig, "POST", execRequest.URL())
	if err != nil {
		return 0, nil
	}
	if !terminal.IsTerminal(0) || !terminal.IsTerminal(1) {
		return 0, err
	}
	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		return 1, err
	}
	defer func(fd int, oldState *terminal.State) error {
		err := terminal.Restore(fd, oldState)
		if err != nil {
			return err
		}
		return nil
	}(0, oldState)

	screen := struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stdout}
	err = exec.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdin:  screen,
		Stdout: screen,
		Stderr: screen,
		Tty:    false,
	})
	var exitCode = 0
	if err != nil {
		if exitErr, ok := err.(utilexec.ExitError); ok && exitErr.Exited() {
			exitCode = exitErr.ExitStatus()
			return 1, nil
		}
	}
	return exitCode, nil
}
func (k *KubernetesApiServiceImpl) CreateRbac(ctx context.Context, serviceAccount *api_v1.ServiceAccount, clusterRole *rbac.ClusterRole, clusterRoleBinding *rbac.ClusterRoleBinding) error {
	_, err := k.clientset.CoreV1().ServiceAccounts("kube-system").Create(ctx, serviceAccount, metav1.CreateOptions{})
	if err != nil {
		if k_error.IsAlreadyExists(err) {
			log.Warnf(err.Error())
		} else {
			return err
		}
	}
	_, err = k.clientset.RbacV1().ClusterRoles().Create(ctx, clusterRole, metav1.CreateOptions{})
	if err != nil {
		if k_error.IsAlreadyExists(err) {
			log.Warnf(err.Error())
		} else {
			return err
		}
	}
	_, err = k.clientset.RbacV1().ClusterRoleBindings().Create(ctx, clusterRoleBinding, metav1.CreateOptions{})
	if err != nil {
		if k_error.IsAlreadyExists(err) {
			log.Warnf(err.Error())
			return nil
		} else {
			return err
		}
	}
	return nil
}

func (k *KubernetesApiServiceImpl) DeployDaemonSet(ctx context.Context, d *apps_v1.DaemonSet) error {
	if _, err := k.clientset.AppsV1().DaemonSets("kube-system").Create(ctx, d, metav1.CreateOptions{}); err != nil {
		if k_error.IsAlreadyExists(err) {
			log.Warnf(err.Error())
		} else {
			return err
		}
	}
	return nil
}

func (k *KubernetesApiServiceImpl) CreateRuntimeClass(ctx context.Context, d *node_v1.RuntimeClass) error {
	if _, err := k.clientset.NodeV1().RuntimeClasses().Create(ctx, d, metav1.CreateOptions{}); err != nil {
		if k_error.IsAlreadyExists(err) {
			log.Warnf(err.Error())
		} else {
			return err
		}
	}
	return nil
}

func (k *KubernetesApiServiceImpl) DeleteDaemonSet(ctx context.Context, d string, namespace string) error {
	err := k.clientset.AppsV1().DaemonSets(namespace).Delete(ctx, d, metav1.DeleteOptions{})
	if err != nil {
		if k_error.IsNotFound(err) {
			log.Warnf(err.Error())
		} else {
			return err
		}
	}
	return nil
}

func (k *KubernetesApiServiceImpl) DeleteRbac() error {
	err := k.clientset.CoreV1().ServiceAccounts("kube-system").Delete(context.TODO(), "kata-label-node", metav1.DeleteOptions{})
	if err != nil {
		if k_error.IsNotFound(err) {
			log.Warnf(err.Error())
		} else {
			return err
		}
	}
	err = k.clientset.RbacV1().ClusterRoles().Delete(context.TODO(), "node-labeler", metav1.DeleteOptions{})
	if err != nil {
		if k_error.IsNotFound(err) {
			log.Warnf(err.Error())
		} else {
			return err
		}
	}
	err = k.clientset.RbacV1().ClusterRoleBindings().Delete(context.TODO(), "kata-label-node-rb", metav1.DeleteOptions{})
	if err != nil {
		if k_error.IsNotFound(err) {
			log.Warnf(err.Error())
		} else {
			return err
		}
	}
	return nil
}

func (k *KubernetesApiServiceImpl) DeleteRuntimeClass(s string) error {
	err := k.clientset.NodeV1().RuntimeClasses().Delete(context.TODO(), s, metav1.DeleteOptions{})
	if err != nil {
		if k_error.IsNotFound(err) {
			log.Warnf(err.Error())
		} else {
			return err
		}
	}
	return nil
}
