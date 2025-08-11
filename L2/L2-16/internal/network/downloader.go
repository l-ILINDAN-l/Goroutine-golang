package network

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func Download(url string, timeoutSec int) (io.ReadCloser, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("error creating request for url %s: %s", url, err)
	}
	req.Header.Set("User-Agent", "wget-downloader")

	client := http.Client{
		Timeout: time.Duration(timeoutSec) * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("error downloading url %s: %s", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("fail download by url %s with status code: %s", url, resp.Status)
	}
	return resp.Body, resp.Header.Get("Content-Type"), nil
}
