package main

import (
	"flag"
	"time"

	"github.com/golang/glog"

	"github.com/nginxinc/kubernetes-ingress/nginx-controller/controller"
	"github.com/nginxinc/kubernetes-ingress/nginx-controller/nginx"
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

var (
	masterURL = flag.String("master", "", `Specify the url of the api server.`)
	masterToken = flag.String("token", "", `Auth token for accessing apiserver.`)
	masterCaFile = flag.String("root-ca-file", "", `Path to root ca file.`)

	watchNamespace = flag.String("watch-namespace", api.NamespaceAll,
		`Namespace to watch for Ingress/Services/Endpoints. By default the controller
		watches acrosss all namespaces`)

	nginxConfigMaps = flag.String("nginx-configmaps", "",
		`Specifies a configmaps resource that can be used to customize NGINX
		configuration. The value must follow the following format: <namespace>/<name>`)
)

func main() {
	flag.Parse()

	var kubeClient *client.Client
	var local = false

	if *masterURL != "" {
		var tlsConfig = &client.TLSClientConfig{
			CAFile: *masterCaFile,
		}

		kubeClient = client.NewOrDie(&client.Config{
			Host: *masterURL,
			BearerToken: *masterToken,
			Insecure:false,
			TLSClientConfig: tlsConfig,
		})
		// local = true
	} else {
		var err error
		kubeClient, err = client.NewInCluster()
		if err != nil {
			glog.Fatalf("Failed to create client: %v.", err)
		}
	}

	glog.V(3).Info("Api client created")
	ngxc, _ := nginx.NewNginxController("/etc/nginx/", local)
	glog.V(3).Info("Starting Nginx")
	ngxc.Start()
	config := nginx.NewDefaultConfig()
	cnf := nginx.NewConfigurator(ngxc, config)
	lbc, _ := controller.NewLoadBalancerController(kubeClient, 30*time.Second, *watchNamespace, cnf, *nginxConfigMaps)
	glog.V(3).Info("Run load balancer controller")
	lbc.Run()
}
