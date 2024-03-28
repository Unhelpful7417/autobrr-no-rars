package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Holds unvalidated POST request data from the user, hence the interfaces. If the
// user submits data that is valid JSON but not valid within the scope of the
// application, we want to return that data in the logs to ease the debugging process.
type InitialRequest struct {
	URL interface{}
	Tolerance interface{}
}

// Fields that are common to all responses
type CommonResponse struct {
	Timestamp  string   `json:"timestamp"`
	Status     int     `json:"status"`
	Method     string  `json:"method"`
	Path       string  `json:"path"`
	Ip         string  `json:"ip"`
	Latency    int64   `json:"latency"`
}

// Fields that are specific to the ValidateTorrentByUrl handler function
type ValidatorResponse struct {
	CommonResponse
	Url        interface{}  `json:"url"`
	Tolerance  interface{}  `json:"tolerance"`
	RarCount   interface{}  `json:"rar_count"`
}

func customLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Keys = make(map[string]any)

		c.Next() // Process request
		latency := time.Since(start)

		var response interface{}
		common := CommonResponse{
			Timestamp: time.Now().Format(time.RFC3339),
			Status:    c.Writer.Status(),
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			Ip:        c.ClientIP(),
			Latency:   latency.Milliseconds(),
		}

		// If we tried to validate a torrent, also include the url
		// and tolerance from the request body in the log output
		if len(c.Keys) != 0 {
			response = ValidatorResponse{
				CommonResponse: common,
				Url: c.Keys["url"],
				Tolerance: c.Keys["tolerance"],
				RarCount: c.Keys["rar_count"],
			}
		} else {
			response = common
		}

		responseJson, err := json.Marshal(response)
		if err != nil { // Should hopefully never occur
			msg := map[string]interface{}{
				"server_error": "error marshaling log data",
				"msg":          err,
				"timestamp":    time.Now().Format(time.RFC3339),
			}
			output, _ := json.Marshal(msg)
			fmt.Println(string(output))
			return
		}

		fmt.Println(string(responseJson))
	}
}

