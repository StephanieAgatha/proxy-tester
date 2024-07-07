package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	goodProxies = make(map[string]bool)
	mutex       sync.Mutex
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
}

func testProxy(proxy string, target string, wg *sync.WaitGroup) {
	defer wg.Done()

	proxyParts := strings.Split(proxy, "@")
	var proxyURL *url.URL
	var err error

	if len(proxyParts) == 2 {
		credentials := proxyParts[0]
		address := proxyParts[1]
		proxyURL, err = url.Parse(fmt.Sprintf("http://%s@%s", credentials, address))
	} else {
		proxyURL, err = url.Parse("http://" + proxy)
	}

	if err != nil {
		log.Error().Err(err).Str("proxy", proxy).Msg("Error parsing proxy")
		return
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 10 * time.Second,
	}

	start := time.Now()
	resp, err := client.Get(target)
	if err != nil {
		if strings.Contains(err.Error(), "Client.Timeout exceeded") {
			log.Warn().Str("proxy", proxy).Str("target", target).Msg("Proxy timed out, skipping")
		} else {
			log.Error().Err(err).Str("proxy", proxy).Str("target", target).Msg("Proxy failed")
		}
		return
	}
	defer resp.Body.Close()

	duration := time.Since(start)
	log.Info().Str("proxy", proxy).Str("target", target).Int64("duration_ms", duration.Milliseconds()).Msg("Proxy succeeded")

	mutex.Lock()
	goodProxies[proxy] = true
	mutex.Unlock()
}

func saveGoodProxies() {
	file, err := os.Create("good_proxy.txt")
	if err != nil {
		log.Error().Err(err).Msg("Error creating good_proxy.txt")
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for proxy := range goodProxies {
		_, err := writer.WriteString(proxy + "\n")
		if err != nil {
			log.Error().Err(err).Msg("Error writing to good_proxy.txt")
			return
		}
	}
	writer.Flush()
	log.Info().Msg("Good proxies saved to good_proxy.txt")
}

func main() {
	file, err := os.Open("proxy.txt")
	if err != nil {
		log.Fatal().Err(err).Msg("Error opening proxy file")
	}
	defer file.Close()

	var wg sync.WaitGroup
	scanner := bufio.NewScanner(file)

	targets := []string{"https://www.youtube.com", "https://www.google.com"}

	for scanner.Scan() {
		proxy := scanner.Text()
		for _, target := range targets {
			wg.Add(1)
			go testProxy(proxy, target, &wg)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Error().Err(err).Msg("Error reading proxy file")
	}

	wg.Wait()
	saveGoodProxies()
	log.Info().Msg("All tests completed")
}
