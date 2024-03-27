package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/anacrolix/torrent/metainfo"
)

// IsValidUrl ensures that a string is a valid URL. Shamelessly stolen from
// https://stackoverflow.com/questions/31480710/validate-url-with-standard-package-in-go.
func IsValidUrl(str string) bool {
    u, err := url.Parse(str)
    return err == nil && u.Scheme != "" && u.Host != "" && (u.Scheme == "http" || u.Scheme == "https")
}

// IsTorrentFile checks if the inputted URL points to a .torrent file after cutting off URL parameters.
func IsTorrentFile(url string) bool {
    trimUrlIndex := strings.Index(url, "?")
    var baseUrl string
    if trimUrlIndex != -1 {
        baseUrl = url[:trimUrlIndex]
    } else {
        baseUrl = url
    }

	return strings.HasSuffix(baseUrl, ".torrent")
}

// Check if port set by user is valid. Checks if the port is an integer between 1-65535.
func isValidPort(portStr string) bool {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return false
	}
	if port >= 1 && port <= 65535 {
		return true
	}
	return false
}

// CheckEnv checks for an environment variable on the system and returns the value
// along with a boolean representing if the environment variable has been set.
func CheckEnv(env string) (value string, isSet bool) {
	value, isSet = os.LookupEnv(env) // Check if $listenPort is set and get value if so
	if isSet {
		return value, true
	} else {
		return "", false
	}
}

// CheckContentLength makes a HEAD request to the provided URL, gets the Content-Length from the response,
// and compares it to the maximum length allowed. One of two possible error values may be returned if the
// HEAD request cannot be completed or if the content length of the resource at the given URL exceeds the tolerance.
func CheckContentLength(client *http.Client, url string, maxLengthInBytes int64) (err error) {
	// Make HEAD request and check length to not fuck up memory
	headResp, err := client.Head(url)
	if err != nil {
		return fmt.Errorf("could not complete HEAD request. error: %v", err)
	}
	if headResp.ContentLength > maxLengthInBytes {
		return fmt.Errorf("content-length of %v exceeds permitted length of %v", headResp.ContentLength, maxLengthInBytes)
	}
	return nil
}

// CheckIfTLCookiesExist looks through a provided cookie jar
// to see if TorrentLeech session cookies have been set.
func CheckIfTLCookiesExist(jar *cookiejar.Jar) bool {
	for _, cookie := range jar.Cookies(tlUrl) {
		if cookie.Name == "tlpass" || cookie.Name == "tluid" {
			return true
		}
	}
	return false
}

// GetFilesFromTorrentInfo gets all the file names contained with the torrent metadata.
func GetFilesFromTorrentInfo(info metainfo.Info) (fileNames []string) {
	if len(info.Files) > 0 {
		for _, file := range info.Files {
			pathList := file.Path // Get full file path like {"folder", "file.txt"}
			fileName := pathList[len(pathList)-1] // Get just file name
			fileNames = append(fileNames, fileName)
		}
	}
	if len(info.Files) == 0 {
		fileNames = append(fileNames, info.Name)
	}
	return fileNames
}

// GetRarFiles searches for strings that look like .rar files or .rar archives,
// i.e., anything like ".rar" or ".r00".
func GetRarFiles(fileNames []string) (rarFileNames []string) {
    // Define the regular expression pattern
    pattern := `\.r\d{2}$`

    // Compile the regular expression
    re := regexp.MustCompile(pattern)
	_ = re
    // Iterate over each string in the slice
    for _, file := range fileNames {
        // Check if the string matches the pattern
        // if re.MatchString(file) || strings.HasSuffix(file, ".rar") {
        if strings.HasSuffix(file, ".rar") {
            rarFileNames = append(rarFileNames, file)
        }
    }
	return rarFileNames
}