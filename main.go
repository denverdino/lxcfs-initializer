// Copyright 2017 Google Inc. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"flag"
	"k8s.io/api/apps/v1"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const (
	defaultAnnotation      = "initializer.kubernetes.io/lxcfs"
	defaultInitializerName = "lxcfs.initializer.kubernetes.io"
	defaultNamespace       = "default"
)

var (
	annotation        string
	configmap         string
	initializerName   string
	namespace         string
	requireAnnotation bool
)

type config struct {
	volumes      []corev1.Volume
	volumeMounts []corev1.VolumeMount
}

func main() {
	flag.StringVar(&annotation, "annotation", defaultAnnotation, "The annotation to trigger initialization")
	flag.StringVar(&initializerName, "initializer-name", defaultInitializerName, "The initializer name")
	flag.StringVar(&namespace, "namespace", "default", "The configuration namespace")
	flag.BoolVar(&requireAnnotation, "require-annotation", true, "Require annotation for initialization")
	flag.Parse()

	log.Println("Starting the Kubernetes initializer...")
	log.Printf("Initializer name set to: %s", initializerName)

	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		log.Fatal(err)
	}

	// -v /var/lib/lxcfs/proc/cpuinfo:/proc/cpuinfo:rw
	// -v /var/lib/lxcfs/proc/diskstats:/proc/diskstats:rw
	// -v /var/lib/lxcfs/proc/meminfo:/proc/meminfo:rw
	// -v /var/lib/lxcfs/proc/stat:/proc/stat:rw
	// -v /var/lib/lxcfs/proc/swaps:/proc/swaps:rw
	// -v /var/lib/lxcfs/proc/uptime:/proc/uptime:rw
	c := &config{
		volumeMounts: []corev1.VolumeMount{
			corev1.VolumeMount{
				Name:      "lxcfs-proc-cpuinfo",
				MountPath: "/proc/cpuinfo",
			},
			corev1.VolumeMount{
				Name:      "lxcfs-proc-meminfo",
				MountPath: "/proc/meminfo",
			},
			corev1.VolumeMount{
				Name:      "lxcfs-proc-diskstats",
				MountPath: "/proc/diskstats",
			},
			corev1.VolumeMount{
				Name:      "lxcfs-proc-stat",
				MountPath: "/proc/stat",
			},
			corev1.VolumeMount{
				Name:      "lxcfs-proc-swaps",
				MountPath: "/proc/swaps",
			},
			corev1.VolumeMount{
				Name:      "lxcfs-proc-uptime",
				MountPath: "/proc/uptime",
			},
		},
		volumes: []corev1.Volume{
			corev1.Volume{
				Name: "lxcfs-proc-cpuinfo",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/var/lib/lxcfs/proc/cpuinfo",
					},
				},
			},
			corev1.Volume{
				Name: "lxcfs-proc-diskstats",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/var/lib/lxcfs/proc/diskstats",
					},
				},
			},
			corev1.Volume{
				Name: "lxcfs-proc-meminfo",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/var/lib/lxcfs/proc/meminfo",
					},
				},
			},
			corev1.Volume{
				Name: "lxcfs-proc-stat",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/var/lib/lxcfs/proc/stat",
					},
				},
			},
			corev1.Volume{
				Name: "lxcfs-proc-swaps",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/var/lib/lxcfs/proc/swaps",
					},
				},
			},
			corev1.Volume{
				Name: "lxcfs-proc-uptime",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/var/lib/lxcfs/proc/uptime",
					},
				},
			},
		},
	}
	// Watch uninitialized Deployments in all namespaces.
	restClient := clientset.AppsV1().RESTClient()
	watchlist := cache.NewListWatchFromClient(restClient, "deployments", corev1.NamespaceAll, fields.Everything())

	// Wrap the returned watchlist to workaround the inability to include
	// the `IncludeUninitialized` list option when setting up watch clients.
	includeUninitializedWatchlist := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.IncludeUninitialized = true
			return watchlist.List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.IncludeUninitialized = true
			return watchlist.Watch(options)
		},
	}

	resyncPeriod := 30 * time.Second

	_, controller := cache.NewInformer(includeUninitializedWatchlist, &v1.Deployment{}, resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				err := initializeDeployment(obj.(*v1.Deployment), c, clientset)
				if err != nil {
					log.Println(err)
				}
			},
		},
	)

	stop := make(chan struct{})
	go controller.Run(stop)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Println("Shutdown signal received, exiting...")
	close(stop)
}

func initializeDeployment(deployment *v1.Deployment, c *config, clientset *kubernetes.Clientset) error {
	if deployment.ObjectMeta.GetInitializers() != nil {
		pendingInitializers := deployment.ObjectMeta.GetInitializers().Pending

		if initializerName == pendingInitializers[0].Name {
			log.Printf("Initializing deployment: %s", deployment.Name)

			initializedDeployment := deployment.DeepCopy()

			// Remove self from the list of pending Initializers while preserving ordering.
			if len(pendingInitializers) == 1 {
				initializedDeployment.ObjectMeta.Initializers = nil
			} else {
				initializedDeployment.ObjectMeta.Initializers.Pending = append(pendingInitializers[:0], pendingInitializers[1:]...)
			}

			if requireAnnotation {
				a := deployment.ObjectMeta.GetAnnotations()
				_, ok := a[annotation]
				if !ok {
					log.Printf("Required '%s' annotation missing; skipping lxcfs injection", annotation)
					_, err := clientset.AppsV1().Deployments(deployment.Namespace).Update(initializedDeployment)
					if err != nil {
						return err
					}
					return nil
				}
			}

			containers := initializedDeployment.Spec.Template.Spec.Containers

			// Modify the Deployment's Pod template to include the Envoy container
			// and configuration volume. Then patch the original deployment.
			for i := range containers {
				containers[i].VolumeMounts = append(containers[i].VolumeMounts, c.volumeMounts...)
			}

			initializedDeployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, c.volumes...)

			oldData, err := json.Marshal(deployment)
			if err != nil {
				return err
			}

			newData, err := json.Marshal(initializedDeployment)
			if err != nil {
				return err
			}

			patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, v1.Deployment{})
			if err != nil {
				return err
			}

			_, err = clientset.AppsV1().Deployments(deployment.Namespace).Patch(deployment.Name, types.StrategicMergePatchType, patchBytes)
			if err != nil {
				return err
			}
		}
	}

	return nil
}