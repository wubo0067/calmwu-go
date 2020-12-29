/*
 * @Author: calm.wu
 * @Date: 2019-08-19 11:31:18
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-08-19 15:11:51
 */

package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		// Set example variable
		c.Set("example", "12345")

		// before request

		c.Next()

		// after request
		latency := time.Since(t)
		log.Print(latency)

		// access the status we are sending
		status := c.Writer.Status()
		log.Println(status)
	}
}

func main() {
	r := gin.New()
	r.Use(Logger())

	r.GET("/test", func(c *gin.Context) {
		example := c.MustGet("example").(string)

		// it would print: "12345"
		log.Println(example)
		c.Writer.WriteHeader(http.StatusOK)
	})

	// :和*的区别，action获得值有/
	r.GET("/user/:name/*action", func(c *gin.Context) {
		name := c.Param("name")
		action := c.Param("action")
		message := name + " is " + action
		log.Println(message)
		c.String(http.StatusOK, message)
	})

	// Listen and serve on 0.0.0.0:8080
	r.Run(":0")
}
