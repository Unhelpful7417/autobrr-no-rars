package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/gin-gonic/gin"
)

var (
	serverPort string // Value of serverPort environment variable
	serverPortIsSet bool
	serverPortIsValid bool
	tlUrl, _ = url.Parse("https://www.torrentleech.org") // Used to do silly TL-specific data checks
	jar, _ = cookiejar.New(nil) // Save cookies between requests to minimize authentication overhead
	client = http.Client{Jar: jar}
)

// Perform data validation
func init() {
	serverPort, serverPortIsSet = CheckEnv("serverPort")
	if serverPortIsSet {
		serverPortIsValid = isValidPort(serverPort) // Check if it's a valid port number
		if !serverPortIsValid {
			log.Printf("WARN: serverPort is invalid. Currently set to: `%s`, will ignore and use default value\n", serverPort)
		}
	}
}

func main() {
	// Initialize Gin router
	gin.SetMode(gin.ReleaseMode) // Turn off release mode warning message
	router := gin.Default()
	router.SetTrustedProxies(nil) // Turn off trusted proxy warning message

	router.POST("/validate-url/", ValidateTorrentByUrl)   // Feels kind of stupid to serve these functions
	router.POST("/validate-url", ValidateTorrentByUrl)    // both with and without the trailing slash but
	router.GET("/healthcheck/", Healthcheck)              // I'm not sure if autobrr follows HTTP301s by
	router.GET("/healthcheck", Healthcheck)               // default - too lazy to test right now and this works
	// router.GET("/get-tl-cookies", getTLSessionCookies)

	if serverPortIsSet && serverPortIsValid {
		serverAddr := fmt.Sprintf("0.0.0.0:%s", serverPort)
		fmt.Printf("Currently listening on %s\n", serverAddr)
		router.Run(serverAddr)
	} else {
		fmt.Println("Currently listening on 0.0.0.0:8080")
		router.Run("0.0.0.0:8080")
	}
}