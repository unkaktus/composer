package composer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

const (
	composeURLBase string = "https://compose.obspm.fr"
)

func downloadPage(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("http.Get: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unsuccessful request: status: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	return string(body), nil
}

func downloadChecksum(u string) (string, error) {
	checksumText, err := downloadPage(u)
	if err != nil {
		return "", fmt.Errorf("download checksum file: %w", err)
	}
	sp := strings.Split(checksumText, " ")
	return sp[0], nil
}

type EOS struct {
	Link         string
	ChecksumLink string
}

func parseEOSPage(page string) (*EOS, error) {
	eos := &EOS{}
	success := false
	htmlTokens := html.NewTokenizer(strings.NewReader(page))
loop:
	for {
		tt := htmlTokens.Next()
		switch tt {
		case html.ErrorToken:
			break loop
		case html.StartTagToken:
			t := htmlTokens.Token()
			isAnchor := t.Data == "a"
			if isAnchor {
				for _, attr := range t.Attr {
					if attr.Key == "href" {
						link, _ := url.Parse(attr.Val)
						base := filepath.Base(link.Path)
						switch base {
						case "eos.zip":
							success = true
							eos.Link = composeURLBase + link.Path
						case "eos.zip_checksum.txt":
							success = true
							eos.ChecksumLink = composeURLBase + link.Path
						}
					}
				}
			}
		}
	}
	if !success {
		return nil, fmt.Errorf("no EOS data found")
	}
	return eos, nil
}

func downloadFile(u, filename string) (string, error) {
	resp, err := http.Get(u)
	if err != nil {
		return "", fmt.Errorf("http.Get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unsuccessful request: status: %s", resp.Status)
	}
	outfile, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer outfile.Close()

	hash := sha256.New()
	mw := io.MultiWriter(outfile, hash)

	_, err = io.Copy(mw, resp.Body)
	if err != nil {
		return "", fmt.Errorf("copy data: %w", err)
	}

	checksum := hex.EncodeToString(hash.Sum(nil))

	return checksum, nil
}

func DownloadEOS(id string) error {
	url := composeURLBase + "/eos/" + id
	page, err := downloadPage(url)
	if err != nil {
		return fmt.Errorf("download page: %w", err)
	}
	eos, err := parseEOSPage(page)
	if err != nil {
		return fmt.Errorf("parse EOS page: %w", err)
	}
	log.Printf("Found EOS data. Downloading.")

	eosDataFilename := fmt.Sprintf("eos_%s.zip", id)
	checksum, err := downloadFile(eos.Link, eosDataFilename)
	if err != nil {
		return fmt.Errorf("download EOS data: %w", err)
	}
	checksumExpected, err := downloadChecksum(eos.ChecksumLink)
	if err != nil {
		return fmt.Errorf("download checksum")
	}
	if checksum != checksumExpected {
		return fmt.Errorf("checksum mismatch")
	}
	return nil
}
