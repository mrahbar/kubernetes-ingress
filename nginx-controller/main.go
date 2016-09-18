package main

import (
	"flag"
	"time"

	"github.com/golang/glog"

	"github.com/nginxinc/kubernetes-ingress/nginx-controller/controller"
	"github.com/nginxinc/kubernetes-ingress/nginx-controller/nginx"
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"

	"k8s.io/client-go/1.4/kubernetes"
	"k8s.io/client-go/1.4/pkg/api"
	"k8s.io/client-go/1.4/tools/clientcmd"
	"fmt"
)

var (
	proxyURL = flag.String("proxy", "",
		`If specified, the controller assumes a kubctl proxy server is running on the
		given url and creates a proxy client. Regenerated NGINX configuration files
    are not written to the disk, instead they are printed to stdout. Also NGINX
    is not getting invoked. This flag is for testing.`)

	kubeConfigPath = flag.String("kubeconfig", "",
		`The nginx controller can use the same kubeconfig file as the
		kubectl CLI does to interact with the apiserver.`)

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

	if *kubeConfigPath != "" {
		// uses the current context in kubeconfig
		config, err := clientcmd.BuildConfigFromFlags("", *kubeConfigPath)
		if err != nil {
			panic(err.Error())
		}
		// creates the clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}
		pods, err := clientset.Core().Pods("").List(api.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	} else {
		if *proxyURL != "" {
			kubeClient = client.NewOrDie(&client.Config{
				Host: *proxyURL,
			})
			local = true
		} else {
			var err error
			kubeClient, err = client.NewInCluster()
			if err != nil {
				glog.Fatalf("Failed to create client: %v.", err)
			}
		}
	}

	ngxc, _ := nginx.NewNginxController("/etc/nginx/", local)
	ngxc.Start()
	config := nginx.NewDefaultConfig()
	cnf := nginx.NewConfigurator(ngxc, config)
	lbc, _ := controller.NewLoadBalancerController(kubeClient, 30*time.Second, *watchNamespace, cnf, *nginxConfigMaps)
	lbc.Run()
}
