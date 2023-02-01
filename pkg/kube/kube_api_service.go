package kube

import (
	"context"
	log "github.com/sirupsen/logrus"
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
