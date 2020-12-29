/*
 * @Author: calm.wu
 * @Date: 2019-09-18 14:46:47
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-09-18 15:16:23
 */

package main

import (
	"reflect"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

func watchNS(clientSet *kubernetes.Clientset) {
	nsWatcher, err := clientSet.CoreV1().Namespaces().Watch(metav1.ListOptions{
		ResourceVersion: "0",
	})
	if err != nil {
		logger.Fatal(err.Error())
	}

	nsWatchChannel := nsWatcher.ResultChan()
	for nsEvent := range nsWatchChannel {
		logger.Printf("event type[%s] object-type:[%s:%T]\n", nsEvent.Type, reflect.TypeOf(nsEvent.Object).String(), nsEvent.Object)
		ns, ok := nsEvent.Object.(*corev1.Namespace)
		if ok {
			switch nsEvent.Type {
			case watch.Added:
				logger.Printf("Add Namespace:%s", ns.Name)
			case watch.Modified:
				logger.Printf("Modified Namespace:%s", ns.Name)
			case watch.Deleted:
				logger.Printf("Deleted Namespace:%s", ns.Name)
			case watch.Bookmark:
				logger.Printf("Bookmark Namespace:%s", ns.Name)
			case watch.Error:
				logger.Printf("Error Namespace:%s", ns.Name)
			}
		} else {
			logger.Fatal("Unexpected type")
		}
	}
}
