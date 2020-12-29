/*
 * @Author: calm.wu
 * @Date: 2019-03-18 18:31:10
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-09-18 15:09:35
 */

package main

import (
	// "k8s.io/kubernetes/pkg/controller/deployment"
	// "k8s.io/kubernetes/pkg/registry/apps/deployment"
	// "k8s.io/kubernetes/pkg/kubectl/util/deployment"
	// "k8s.io/kubernetes/pkg/api/v1/service"
	"log"
	"os"
	"reflect"

	"github.com/fatih/color"
	"github.com/urfave/cli"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/helm/pkg/helm"
	//"k8s.io/apimachinery/pkg/labels"
)

var (
	logger = log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)
)

func listPod(clientSet *kubernetes.Clientset) {
	pods, err := clientSet.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		logger.Fatal(err)
	}
	for i, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodPending {
			color.New(color.FgBlue).Printf("\t{%d}: pod:%s status:%s PodIP:%s\n", i, pod.Name, pod.Status.Phase, pod.Status.PodIP)
		} else {
			logger.Printf("\t{%d}: pod:%s status:%s PodIP:%s\n", i, pod.Name, pod.Status.Phase, pod.Status.PodIP)
		}
	}
}

func listDeployment(clientSet *kubernetes.Clientset) {
	// 由于我的yam文件中写的是apiVersion: extensions/v1beta1，所以这里使用ExtensionsV1beta1来查询，这点很重要
	deploymentsClient := clientSet.ExtensionsV1beta1().Deployments(corev1.NamespaceDefault)
	//deploymentsClient := clientSet.AppsV1().Deployments(corev1.NamespaceDefault)

	deployments, err := deploymentsClient.List(metav1.ListOptions{})
	if err != nil {
		logger.Fatal(err)
	}

	for i, deployment := range deployments.Items {
		logger.Printf("\t{%d}: deployment:%s (%d replicas)\n", i, deployment.Name, *deployment.Spec.Replicas)
	}
}

func listServices(clientSet *kubernetes.Clientset) {
	var kubeClient kubernetes.Interface = clientSet
	servicesClient := kubeClient.CoreV1().Services(corev1.NamespaceDefault)

	services, err := servicesClient.List(metav1.ListOptions{})
	if err != nil {
		logger.Fatal(err)
	}

	for i := range services.Items {
		service := &services.Items[i]
		logger.Printf("\t{%d}: service:%s labes:%#v\n", i, service.Name, service.Labels)
	}
}

func createJob(clientSet *kubernetes.Clientset) {
	var activeDeadlineSecs int64 = 50 // 活动超时时间
	var parallelism int32 = 2         // 同时启动pod的数量
	var jobTTL int32 = 60             // 完毕后存活时间

	jobClient := clientSet.BatchV1().Jobs("default")
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sleep-job",
			Labels: map[string]string{
				"app": "pci-sleep-job",
			},
		},
		Spec: batchv1.JobSpec{
			Parallelism: &parallelism,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "sleep-job",
							Image:   "busybox",
							Command: []string{"sleep", "30"},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			ActiveDeadlineSeconds:   &activeDeadlineSecs, // 如果job运行完毕，这个是没有作用的
			TTLSecondsAfterFinished: &jobTTL,             // 执行完毕后等待多久后删除，项目的目标不是自动删除job
		},
	}

	// 创建job,
	// job无法重复创建
	// 删除了job，job下面的pod也被释放
	jobRes, err := jobClient.Create(job)
	if err != nil {
		logger.Fatalf("create sleep-job failed, reason:%s\n", err.Error())
	}

	logger.Printf("sleep-job:%s\n", jobRes.GetObjectMeta().GetName())

	// 开始watch
	//labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"app":"pci-sleep-job"}}
	jobWatcher, _ := jobClient.Watch(metav1.ListOptions{
		ResourceVersion: "0",
		LabelSelector:   "app=pci-sleep-job",
		// LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	})

	for e := range jobWatcher.ResultChan() {
		logger.Printf("event type[%s] object-type:[%s:%T]\n",
			e.Type, reflect.TypeOf(e.Object).String(), e.Object)
		job, ok := e.Object.(*batchv1.Job)
		if ok {
			logger.Printf("job:%s status:%s", job.Name, job.Status.String())
			if len(job.Status.Conditions) > 0 {
				jobCond := job.Status.Conditions[0]
				if jobCond.Type == batchv1.JobFailed {
					logger.Printf("---job:%s failed, so delete----\n", job.Name)
					// 删除job，job下的pod会被删除
					// err := jobClient.Delete(job.Name, &metav1.DeleteOptions{})
					// if err != nil {
					// 	logger.Fatalf("delete job:%s failed:%s\n", job.Name, err.Error())
					// }
				} else if jobCond.Type == batchv1.JobComplete {
					logger.Printf("---job:%s completed, so delete----\n", job.Name)
					// 这里只会删除job，job下的pod不会被删除
					// err := jobClient.Delete(job.Name, &metav1.DeleteOptions{})
					// if err != nil {
					// 	logger.Fatalf("delete job:%s failed:%s\n", job.Name, err.Error())
					// }
				}
			}
		}
	}
}

func createHelmClient(tillerHost string) {
	helm.NewClient(helm.Host(tillerHost))
}

func testServiceSpec() {
	svcSpec := corev1.ServiceSpec{}
	logger.Printf("svcSpec:#v\n", svcSpec)
}

func main() {
	app := cli.NewApp()
	app.Name = "k8sclient"
	app.Usage = "k8sclient"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "kubeconfig",
			Value: "",
			Usage: "kubeconfig",
		},
	}

	app.Action = func(c *cli.Context) error {
		kubeconfig := c.String("kubeconfig")
		// 判断文件是否存在
		_, err := os.Stat(kubeconfig)
		if err != nil {
			logger.Fatal(err)
		}

		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			logger.Fatal(err)
		}

		clientSet, err := kubernetes.NewForConfig(config)
		if err != nil {
			logger.Fatal(err)
		}

		//listPod(clientSet)
		//listDeployment(clientSet)
		//listServices(clientSet)
		//createJob(clientSet)
		watchNS(clientSet)
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Printf("k8sclient exit!")
}
