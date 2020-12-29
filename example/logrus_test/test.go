/*
 * @Author: calm.wu
 * @Date: 2019-04-24 16:52:31
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-04-24 17:07:08
 */

package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	logFile, _ := os.OpenFile("logrus.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	defer logFile.Close()

	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(logFile)

	logrus.WithFields(logrus.Fields{
		"animal": "walrus",
	}).Info("A walrus appears")
}
