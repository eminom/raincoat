package metadata

import (
	"encoding/json"
	"log"
	"os"
)

// map[string]map[string][]string

type JsonLoader struct{}

func (JsonLoader) LoadInfo(filename string) map[string]map[string][]string {
	chunk, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("Could not open %v for read", filename)
		return nil
	}
	var dict map[string]map[string][]string
	err = json.Unmarshal(chunk, &dict)
	if err != nil {
		log.Printf("unmarshall error: %v", err)
		return nil
	}
	return dict
}

func LoadSubOpsJson(filename string) map[string]map[string][]string {
	return JsonLoader{}.LoadInfo(filename)
}
