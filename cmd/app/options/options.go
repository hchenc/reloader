package options

import (
	"errors"
	"fmt"
	"io/ioutil"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/homedir"
	"net"
	"os"
	"os/user"
	"path"
)

var (
	LeaderElect bool
)

type KubernetesOptions struct {
	// kubeconfig
	KubeConfig *rest.Config

	// kubeconfig path, if not specified, will use
	// in cluster way to create clientset
	KubeConfigPath string `json:"kubeconfig" yaml:"kubeconfig"`

	// kubernetes apiserver public address, used to generate kubeconfig
	// for downloading, default to host defined in kubeconfig
	// +optional
	Master string `json:"master,omitempty" yaml:"master"`

	// kubernetes clientset qps
	// +optional
	QPS float32 `json:"qps,omitempty" yaml:"qps"`

	// kubernetes clientset burst
	// +optional
	Burst int `json:"burst,omitempty" yaml:"burst"`
}

// NewKubernetesOptions returns a `zero` instance
func NewKubernetesConfig() (option *KubernetesOptions) {
	option = &KubernetesOptions{
		QPS:   1e6,
		Burst: 1e6,
	}

	// make it be easier for those who wants to run api-server locally
	homePath := homedir.HomeDir()
	if homePath == "" {
		// try os/user.HomeDir when $HOME is unset.
		if u, err := user.Current(); err == nil {
			homePath = u.HomeDir
		}
	}

	userHomeConfig := path.Join(homePath, ".kube/config")
	if _, err := os.Stat(userHomeConfig); err == nil {
		option.KubeConfigPath = userHomeConfig
	}
	return
}

func (k *KubernetesOptions) Validate() []error {
	var errs []error

	if len(k.KubeConfigPath) != 0 {
		if config, err := clientcmd.BuildConfigFromFlags("", k.KubeConfigPath); err == nil {
			k.KubeConfig = config
		} else {
			errs = append(errs, err)
			return errs
		}
	} else {
		const (
			tokenFile  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
			rootCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
		)
		host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
		if len(host) == 0 || len(port) == 0 {
			errs = append(errs, errors.New("unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined"))
			return errs
		}

		token, err := ioutil.ReadFile(tokenFile)
		if err != nil {
			errs = append(errs, err)
			return errs
		}
		tlsClientConfig := rest.TLSClientConfig{}
		if _, err := certutil.NewPool(rootCAFile); err != nil {
			errs = append(errs, fmt.Errorf("expected to load root CA config from %s, but got err: %v", rootCAFile, err))
			return errs
		} else {
			tlsClientConfig.CAFile = rootCAFile
		}

		k.KubeConfig = &rest.Config{
			// TODO: switch to using cluster DNS.
			Host:            "https://" + net.JoinHostPort(host, port),
			TLSClientConfig: tlsClientConfig,
			BearerToken:     string(token),
			BearerTokenFile: tokenFile,
		}
	}
	return errs
}
