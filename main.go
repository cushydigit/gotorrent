package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: gotorrent <infohash>")
		return
	}
	infohash := os.Args[1]

	downloadDir := filepath.Join("wd", "downloads")

	// Create the download directory if it doesn't exist
	if err := os.MkdirAll(downloadDir, os.ModePerm); err != nil {
		fmt.Println("Error creating download directory:", err)
		return
	}
	// Create a new torrent client
	clientConfig := torrent.NewDefaultClientConfig()
	clientConfig.DataDir = downloadDir
	client, err := torrent.NewClient(clientConfig)
	if err != nil {
		fmt.Println("Error creating client", err)
	}
	defer client.Close()
	// Convert the infohash string to a metainfo.Hash
	hash := metainfo.NewHashFromHex(infohash)

	// Add the torrent using the infohash
	torrent, ok := client.AddTorrentInfoHash(hash)
	if !ok {
		fmt.Println("Error adding torrent:", err)
		return
	}
	<-torrent.GotInfo()
	fmt.Println("Torrent info loaded:", torrent.Name())

	torrent.DownloadAll()

	// handle gracefull shudown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		fmt.Println("Shutting down...")
		client.Close()
		os.Exit(0)
	}()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	tSize := 40

	for {
		select {
		case <-ticker.C:
			completed := torrent.BytesCompleted()
			total := torrent.Info().TotalLength()

			// Print download progress
			progress := float64(completed) / float64(total) * 100

			// Draw progress bar
			filled := int(progress / 100 * float64(tSize))
			bar := fmt.Sprintf("%s%s", strings.Repeat("█", filled), strings.Repeat("░", tSize-filled))
			fmt.Printf("\r%s %.2f%%", bar, progress)
			os.Stdout.Sync()

			if completed == total {
				fmt.Printf("\nDownload completed!\n")
				return
			}

		case <-torrent.Closed():
			fmt.Println("Torrent closed")
			return
		}
	}

}
