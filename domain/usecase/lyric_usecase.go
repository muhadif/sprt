package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// LyricUseCase defines the interface for lyric-related use cases.
type LyricUseCase interface {
	// GetLyrics retrieves the lyrics for the given artist, title, and album.
	GetLyrics(ctx context.Context, artist, title, album string) (*Lyrics, error)
	DisplaySyncedLyrics(ctx context.Context, lyrics *Lyrics, startTimeMs int, playerUseCase PlayerUseCase)
}

// Lyrics represents a song's lyrics with timing information.
type Lyrics struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	Language string `json:"language"`
	Synced   bool   `json:"syncedLyrics"`
	Lines    []Line `json:"lines"`
}

// Line represents a single line of lyrics with timing information.
type Line struct {
	StartTimeMs int    `json:"startTimeMs"`
	EndTimeMs   int    `json:"endTimeMs"`
	Text        string `json:"text"`
}

// lyricUseCase implements the LyricUseCase interface.
type lyricUseCase struct {
	cache     map[string]*Lyrics
	cacheLock sync.RWMutex
}

// NewLyricUseCase creates a new instance of LyricUseCase.
func NewLyricUseCase() LyricUseCase {
	return &lyricUseCase{
		cache: make(map[string]*Lyrics),
	}
}

// GetLyrics retrieves the lyrics for the given artist, title, and album.
func (l *lyricUseCase) GetLyrics(ctx context.Context, artist, title, album string) (*Lyrics, error) {
	// Create a cache key from artist and title
	cacheKey := artist + "|" + title

	// Check if lyrics are in the cache
	l.cacheLock.RLock()
	cachedLyrics, found := l.cache[cacheKey]
	l.cacheLock.RUnlock()

	if found {
		fmt.Println("Using cached lyrics for", title, "by", artist)
		return cachedLyrics, nil
	}

	// Lyrics not in cache, fetch from API
	// Prepare the request to lrclib.net
	baseURL := "https://lrclib.net/api/search"
	params := url.Values{}
	params.Set("track_name", title)
	params.Set("artist_name", artist)

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get lyrics: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	type libResponse struct {
		Id           int     `json:"id"`
		Name         string  `json:"name"`
		TrackName    string  `json:"trackName"`
		ArtistName   string  `json:"artistName"`
		AlbumName    string  `json:"albumName"`
		Duration     float32 `json:"duration"`
		Instrumental bool    `json:"instrumental"`
		PlainLyrics  *string `json:"plainLyrics"`
		SyncedLyrics *string `json:"syncedLyrics"`
	}

	var libResponses []libResponse
	if err := json.Unmarshal(body, &libResponses); err != nil {
		fmt.Println("err", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if lyrics were found
	if len(libResponses) == 0 {
		return nil, fmt.Errorf("no lyrics found for %s by %s", title, artist)
	}

	// Find the first synced lyrics if available
	var selectedLyrics *libResponse
	for i := range libResponses {
		if libResponses[i].SyncedLyrics != nil {
			selectedLyrics = &libResponses[i]
			break
		}
	}
	if selectedLyrics == nil {
		selectedLyrics = &libResponses[0]
	}

	// Parse the synced lyrics
	lyrics := &Lyrics{
		ID:     selectedLyrics.Id,
		Name:   selectedLyrics.Name,
		Artist: selectedLyrics.ArtistName,
		Album:  selectedLyrics.AlbumName,
		Synced: selectedLyrics.SyncedLyrics != nil,
		Lines:  []Line{},
	}

	if selectedLyrics.SyncedLyrics != nil {
		// Parse the LRC format
		lines := strings.Split(*selectedLyrics.SyncedLyrics, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}

			// Parse the timestamp and text
			// Format: [mm:ss.xx]text
			if !strings.HasPrefix(line, "[") {
				continue
			}

			closeBracket := strings.Index(line, "]")
			if closeBracket == -1 {
				continue
			}

			timestamp := line[1:closeBracket]
			text := line[closeBracket+1:]

			// Parse the timestamp
			var minutes, seconds, milliseconds int
			if _, err := fmt.Sscanf(timestamp, "%d:%d.%d", &minutes, &seconds, &milliseconds); err != nil {
				continue
			}

			// Convert to milliseconds
			startTimeMs := minutes*60*1000 + seconds*1000 + milliseconds*10

			// Add the line
			lyrics.Lines = append(lyrics.Lines, Line{
				StartTimeMs: startTimeMs,
				EndTimeMs:   0, // Will be set below
				Text:        text,
			})
		}

		// Set the end time for each line
		for i := 0; i < len(lyrics.Lines)-1; i++ {
			lyrics.Lines[i].EndTimeMs = lyrics.Lines[i+1].StartTimeMs
		}
		if len(lyrics.Lines) > 0 {
			// Set a default end time for the last line
			lyrics.Lines[len(lyrics.Lines)-1].EndTimeMs = lyrics.Lines[len(lyrics.Lines)-1].StartTimeMs + 5000
		}
	}

	// Store lyrics in cache
	l.cacheLock.Lock()
	l.cache[cacheKey] = lyrics
	l.cacheLock.Unlock()

	return lyrics, nil
}

// DisplaySyncedLyrics displays the lyrics synchronized with the music.
// It polls Spotify every 3 seconds to keep the lyrics in sync with the currently playing track.
func (l *lyricUseCase) DisplaySyncedLyrics(ctx context.Context, lyrics *Lyrics, startTimeMs int, playerUseCase PlayerUseCase) {
	if lyrics == nil || len(lyrics.Lines) == 0 {
		fmt.Println("No lyrics to display.")
		return
	}

	// Find the current line based on the start time
	currentLineIndex := 0
	for i, line := range lyrics.Lines {
		if line.StartTimeMs <= startTimeMs && startTimeMs < line.EndTimeMs {
			currentLineIndex = i
			break
		} else if line.StartTimeMs > startTimeMs {
			break
		}
		currentLineIndex = i
	}

	// Create a ticker to poll Spotify every 3 seconds
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Display the lyrics synchronized with the music
	startTime := time.Now().Add(-time.Duration(startTimeMs) * time.Millisecond)
	currentProgressMs := startTimeMs

	// Track the last update time to prevent too frequent updates (blinking)
	lastUpdateTime := time.Now().Add(-2 * time.Second) // Initialize to allow immediate first update

	// Create a channel to signal when we need to update the display
	updateCh := make(chan struct{}, 1)
	// Initial update
	updateCh <- struct{}{}

	// Start a goroutine to poll Spotify
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Get the currently playing track
				track, err := playerUseCase.GetCurrentlyPlayingDetails(ctx)
				if err != nil {
					fmt.Printf("\rError getting track: %v", err)
					continue
				}

				if track.Title != lyrics.Name {
					fmt.Printf("\rLyrics for %s by %s", track.Title, track.Artist)
					lyrics, err = l.GetLyrics(ctx, track.Artist, track.Title, track.Album)
					if err != nil {
						fmt.Printf("\rError getting lyrics: %v", err)
						continue
					}
				}

				// Update the progress and signal for display update
				currentProgressMs = track.ProgressMs
				startTime = time.Now().Add(-time.Duration(currentProgressMs) * time.Millisecond)

				// Signal for update
				select {
				case updateCh <- struct{}{}:
				default:
					// Channel already has an update pending
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-updateCh:
			// Check if enough time has passed since the last update (1 second)
			// This prevents the blinking effect caused by too frequent updates
			if time.Since(lastUpdateTime) < time.Second {
				continue // Skip this update if it's too soon
			}
			lastUpdateTime = time.Now() // Update the last update time

			// Find the current line based on the current progress
			currentLineIndex = 0
			for i, line := range lyrics.Lines {
				if line.StartTimeMs <= currentProgressMs && currentProgressMs < line.EndTimeMs {
					currentLineIndex = i
					break
				} else if line.StartTimeMs > currentProgressMs {
					break
				}
				currentLineIndex = i
			}

			// Display the current line
			if currentLineIndex < len(lyrics.Lines) {
				line := lyrics.Lines[currentLineIndex]
				// Clear the line and display the current lyric
				fmt.Print("\r\033[K") // Clear the line
				fmt.Print(line.Text)

				// Write the current line to a file for external use
				err := os.WriteFile("/tmp/current-lyric.txt", []byte(line.Text), 0644)
				if err != nil {
					fmt.Printf("\nError writing to file: %v", err)
				}

				// Calculate when to display the next line
				if currentLineIndex < len(lyrics.Lines)-1 {
					nextLine := lyrics.Lines[currentLineIndex+1]
					waitTime := time.Until(startTime.Add(time.Duration(nextLine.StartTimeMs) * time.Millisecond))

					// Set a timer to update when it's time for the next line
					if waitTime > 0 {
						time.AfterFunc(waitTime, func() {
							select {
							case updateCh <- struct{}{}:
							default:
								// Channel already has an update pending
							}
						})
					} else {
						// If we're already past the next line's start time, update immediately
						go func() {
							updateCh <- struct{}{}
						}()
					}
				}
			}
		}
	}
}
