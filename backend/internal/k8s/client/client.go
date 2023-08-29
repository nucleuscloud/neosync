package neosync_k8sclient

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
)

type Client struct {
	K8sClient            *kubernetes.Clientset
	CustomResourceClient runtimeclient.WithWatch
}

func New() (*Client, error) {
	cfg, err := getRestConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to get InClusterConfig: %w", err)
	}
	return newFromConfig(cfg)
}

func getRestConfig() (*rest.Config, error) {
	// can update this if we ever want to run the backend outside of the cluster to pull from $KUBECONFIG or $HOME/.kube/config
	return rest.InClusterConfig()
}

func newFromConfig(
	cfg *rest.Config,
) (*Client, error) {
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to init kubernetes clientset: %w", err)
	}

	scheme, err := buildCustomResourceScheme()
	if err != nil {
		return nil, fmt.Errorf("unable to build custom resource scheme: %w", err)
	}

	rtclient, err := runtimeclient.NewWithWatch(cfg, runtimeclient.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("unable to init kubernetes runtime client: %w", err)
	}

	return &Client{
		K8sClient:            clientset,
		CustomResourceClient: rtclient,
	}, nil
}

func buildCustomResourceScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()

	err := neosyncdevv1alpha1.AddToScheme(scheme)
	if err != nil {
		return nil, fmt.Errorf("unable to register neosync.dev v1alpha1: %w", err)
	}

	return scheme, nil
}
