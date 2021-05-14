/*
 * @Description: Do not edit
 * @Author: Jianxuesong
 * @Date: 2021-05-13 16:15:11
 * @LastEditors: Jianxuesong
 * @LastEditTime: 2021-05-14 14:16:38
 * @FilePath: /Coco/main.go
 */
package main

import (
	"io"
	"log"
	"os"

	lf "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

func main() {
	// log := logrus.New()
	logrus.SetReportCaller(true)
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetFormatter(&lf.Formatter{
		HideKeys:        true,
		TimestampFormat: "[2006/01/02 15:04:05]",
		// FieldsOrder:     []string{"name", "age"},
		ShowFullLevel: true,
	})

	stdout := os.Stdout
	file, err := os.OpenFile("/tmp/error.log", os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		log.Fatalf("create file log.txt failed: %v", err)
	}

	logrus.SetOutput(io.MultiWriter(file, stdout))

	logrus.WithField("component", "main").Error("sdf")
	// logrus.Trace("trace msg")
	// logrus.Debug("debug msg")
	// logrus.Info("info msg")
	// logrus.Warn("warn msg")
	// logrus.Error("error msg")

}
