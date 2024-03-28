package main

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/gin-gonic/gin"
)

// Ensures a URL submitted by the user points to a .torrent file that does not contain .rar files
func ValidateTorrentByUrl(c *gin.Context) {
	// Ensure data from user is JSON
	var init InitialRequest
	if err := c.ShouldBindJSON(&init); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request schema",
			"msg": err.Error(),
		})
		return
	}
	c.Keys["url"] = init.URL
	c.Keys["tolerance"] = init.Tolerance

	// Ensure data from the user can be parsed
	reqUrl, err := assertToString(c.Keys["url"])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid url value",
			"url": c.Keys["url"],
		})
		return
	}
	reqTolerance, err := assertToValidInt(c.Keys["tolerance"])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid tolerance value, should be 0-255",
			"tolerance": c.Keys["tolerance"],
		})
		return
	}

	// Check if user has submitted a valid URL
	if !IsValidUrl(reqUrl) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid url",
			"url": reqUrl,
		})
		return
	}

	// Trim off extra URL parameters and make sure we're trying to download a torrent file
	if !IsTorrentFile(reqUrl) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "url does not point to a .torrent file",
			"url": reqUrl,
		})
		return
	}

	// Make HEAD request to check length so as not to waste memory
	if err := CheckContentLength(&client, reqUrl, 100000000); err != nil { // 100MB size limit on torrent file
		c.JSON(http.StatusBadRequest, gin.H{"error": "content length check failed"})
		return
	}

	// Create HTTP request to download .torrent file
	httpReq, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not create GET request"})
		return
	}


	// Check if torrent file is on TL
	urlcheck, err := url.Parse(reqUrl)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "could not parse url",
			"url": reqUrl,
		})
		return
	}
	
	// Need to log in to TorrentLeech to download valid torrent files. If we don't yet have session cookies from TL, then
	// we POST the username and password to the landing page to get session cookies. These cookies are saved between runs
	// of this function, so if we already have a session cookie for TL then we ignore this check and use the cookie jar. 
	torrentIsFromTL := strings.HasSuffix(urlcheck.Hostname(), "torrentleech.org")
	if torrentIsFromTL && !CheckIfTLCookiesExist(jar) {
		// Get username and password from environment variables
		tlUsername, isSet := CheckEnv("tlUsername")
		if !isSet {
			log.Printf("ERROR: environment variable tlUsername not set, cannot check torrent at %v\n", reqUrl)
		}
		tlPassword, isSet := CheckEnv("tlPassword")
		if !isSet {
			log.Printf("ERROR: environment variable tlPassword not set, cannot check torrent at %v\n", reqUrl)
		}
		loginTL := url.Values{
			"username": {tlUsername},
			"password": {tlPassword},
		}

		// Submit a simple POST request to the landing page to get some session cookies. Since we're
		// using the http.client/cookie jar declared with global scope, the cookies are saved globally.
		// This limits the number of authentication requests we send to TL and makes us less spammy
		resp, err := client.PostForm(reqUrl, loginTL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "could not authenticate to TL",
				"url": reqUrl,
			})
			return
		}
		defer resp.Body.Close()
	}

	// Execute request
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error downloading torrent file"})
		return
	}
	defer resp.Body.Close()

	// Read the downloaded .torrent file into memory
	var fileBuf bytes.Buffer
	_, err = io.Copy(&fileBuf, resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error reading torrent file"})
		return
	}

	// Parse the downloaded .torrent file from memory
	mi, err := metainfo.Load(bytes.NewReader(fileBuf.Bytes()))
	// Do some unique error handling for TorrentLeech since there's some unique jank with them. When downloading a torrent from them without
	// authentication, their server will still provide a torrent file but it's invalid - when trying to parse it, will return an error like:
	// bencode: syntax error (offset: 0): unknown value type '\u003c'
	// To provide the user with a helpful error, we check if we get a bencode syntax error from TL. If so, then
	// assume it's due to an unauthenticated request because the user did not provide valid credentials.
	var e *bencode.SyntaxError
	if err != nil && torrentIsFromTL && errors.As(err, &e) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "torrentleech credentials invalid",
		})
		return
	}
	// Generic error handling
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error parsing torrent file", "msg": err.Error()})
		return
	}

	// Get metadata from .torrent file
	info, err := mi.UnmarshalInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot unmarshal torrent metadata"})
		return
	}

	// Get file names from torrent info
	fileNames := GetFilesFromTorrentInfo(info)
	// Filter to find ones that look like .rar archive files
	rarFileNames := GetRarFiles(fileNames)
	c.Keys["rar_count"] = len(rarFileNames)

	// Perform final check for .rar files
	if len(rarFileNames) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"msg": "torrent is free of rar archives",
			"url": reqUrl,
			"tolerance": reqTolerance,
		})
		return
	}
	if len(rarFileNames) > int(reqTolerance) && reqTolerance == 0 {
		c.JSON(http.StatusTeapot, gin.H{
			"msg": "rar files found in torrent metadata",
			"url": reqUrl,
			"tolerance": reqTolerance,
			"rar_files": rarFileNames,
		})
		return
	}
	if len(rarFileNames) <= int(reqTolerance) && reqTolerance != 0 {
		c.JSON(http.StatusOK, gin.H{
			"msg": "rar files were found but count within tolerance",
			"url": reqUrl,
			"tolerance": reqTolerance,
			"rar_files": rarFileNames,
		})
		return
	}
	
	// This should never happen
	c.JSON(http.StatusInternalServerError, gin.H{
		"msg": "logic failure, user did something unexpected - submit an issue please :)",
	})
}

// getTLSessionCookies returns the session cookies associated with TorrentLeech. Used for debugging.
func getTLSessionCookies(c *gin.Context) {
	var output []*http.Cookie
	for _, cookie := range client.Jar.Cookies(tlUrl) {
		if cookie.Name == "tluid" || cookie.Name == "tlpass" {
			output = append(output, cookie)
		}
	}
	c.JSON(http.StatusAccepted, gin.H{
		"cookies": output,
	})
}

// Returns a simple message as a health check
func Healthcheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "okay",
	})
}