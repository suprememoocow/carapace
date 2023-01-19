package list

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/itchyny/gojq"
)

type (
	shellyWithSettings struct {
		HostName string      `json:"hostname"`
		Address  string      `json:"address"`
		Settings interface{} `json:"settings"`
	}
)

func QueryShellies(jqQuery string, timeout time.Duration) error {
	if jqQuery == "" { 
		jqQuery =  "."
	}

	query, err := gojq.Parse(jqQuery)
	if err != nil {
		log.Fatalln(err)
	}

	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return fmt.Errorf("failed to initialize resolver: %w", err)
	}

	var wg sync.WaitGroup

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			if isShelly(entry) {
				wg.Add(1)
				go func() {
					err := queryShelly(query, entry, &wg)
					if err != nil {
						log.Printf("Failed to query: %v", err)
					}
				}()
			}
		}
	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err = resolver.Browse(ctx, "_http._tcp", "local", entries)
	if err != nil {
		log.Fatalln("Failed to browse:", err.Error())
	}

	<-ctx.Done()
	wg.Wait()

	return nil
}

func makeGetRequest(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// if password != "" {
	// 	req.SetBasicAuth(username, password)
	// }

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func queryShelly(query *gojq.Query, entry *zeroconf.ServiceEntry, wg *sync.WaitGroup) error {
	defer wg.Done()

	shellyAddress := entry.AddrIPv4[0].String()
	s := shellyWithSettings{
		HostName: entry.HostName,
		Address:  shellyAddress,
	}

	u := url.URL{}
	u.Scheme = "http"
	u.Host = shellyAddress
	u.Path = "settings"

	body, err := makeGetRequest(u.String())
	if err != nil {
		return fmt.Errorf("failed to query %s (%s): %w", entry.HostName, shellyAddress, err)
	}

	var settings map[string]interface{}

	err = json.Unmarshal(body, &settings)
	if err != nil {
		return err
	}
	
	s.Settings = settings

	entryJSONBytes, err := json.Marshal(s)
	if err != nil {
		return err
	}

	// This is very inefficient, but doesn't need to 
	// scale
	var settingsWithShellyMap map[string]interface{}
	
	err = json.Unmarshal(entryJSONBytes, &settingsWithShellyMap)
	if err != nil {
		return err
	}

	iter := query.Run(settingsWithShellyMap)
	
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return fmt.Errorf("jq query failed: %w", err)
		}

		if vAsString, ok := v.(string); ok {
			// If result is a string, emit it raw
			fmt.Println(vAsString)
			continue
		}

		entryJSONBytes, err := json.Marshal(v)
		if err != nil {
			return err
		}
		fmt.Println(string(entryJSONBytes))
	}

	return nil
}

func isShelly(entry *zeroconf.ServiceEntry) bool {
	for _, t := range entry.Text {
		if t == "arch=esp8266" {
			return true
		}
	}
	return false
}
