package main

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func fetchImages(ctx context.Context, url string) ([]string, error) {
	var nodes []*cdp.Node

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Nodes(`img`, &nodes, chromedp.ByQueryAll),
	)
	if err != nil {
		return nil, err
	}

	var imageURLs []string
	for _, n := range nodes {
		for i := 0; i < len(n.Attributes); i += 2 {
			if n.Attributes[i] == "src" {
				imageURLs = append(imageURLs, n.Attributes[i+1])
			}
		}
	}

	return imageURLs, nil
}

func downloadImage(imageURL, path string) error {
	resp, err := http.Get(imageURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func main() {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	dateStr := time.Now().Format("20060102")
	downloadDir := filepath.Join(os.Getenv("HOME"), "Downloads", dateStr) // Change to your desired directory

	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		panic(err)
	}

	baseURL := "https://blog.goo.ne.jp/monsteraquarium/images/?p="
	for i := 1; i <= 10; i++ {
		url := fmt.Sprintf("%s%d", baseURL, i)
		images, err := fetchImages(ctx, url)
		if err != nil {
			fmt.Printf("Error fetching images from %s: %v\n", url, err)
			continue
		}

		// Skip first 16 images
		if len(images) > 16 {
			images = images[16:]
		} else {
			continue
		}

		for j, img := range images {
			err = downloadImage(img, filepath.Join(downloadDir, fmt.Sprintf("image%d_%d.jpg", i, j+1)))
			if err != nil {
				fmt.Printf("Error downloading image %s: %v\n", img, err)
			}
		}
	}
}
