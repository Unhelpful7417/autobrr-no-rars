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

	// Define route to handle torrent file download and parsing
	router.POST("/validate-url", ValidateTorrentByUrl)
	// router.GET("/get-tl-cookies", getTLSessionCookies)
	router.GET("/healthcheck", Healthcheck)

	// Listen on any address using custom port if set by user, defaulting to port 8080 if not set
	if serverPortIsSet && serverPortIsValid {
		serverAddr := fmt.Sprintf("0.0.0.0:%s", serverPort)
		fmt.Printf("Currently listening on %s\n", serverAddr)
		router.Run(serverAddr)
	} else {
		fmt.Println("Currently listening on 0.0.0.0:8080")
		router.Run("0.0.0.0:8080")
	}
}