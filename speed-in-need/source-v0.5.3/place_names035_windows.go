package main

import (
	"encoding/json"
	"strings"
	"sync"
	"unicode"
)

type placeNameMapping035 struct{ Source, Display string }

var placeNameMappings035 []placeNameMapping035
var placeNameMappingLoad035 sync.Once

func placeTitle035(name string) string {
	words := strings.Fields(strings.ToLower(name))
	for i, word := range words {
		runes := []rune(word)
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
		}
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

// Only the displayed cartographic name changes. The canonical mapping is
// shared with rebuild035_maps.py; route definitions keep their original names.
func friendlyPlaceName035(name string) string {
	placeNameMappingLoad035.Do(func() {
		if raw, err := assets.ReadFile("assets/b037_place_names.json"); err == nil {
			json.Unmarshal(raw, &placeNameMappings035)
		}
	})
	upper := strings.ToUpper(strings.TrimSpace(name))
	for _, mapping := range placeNameMappings035 {
		if upper == mapping.Source {
			return mapping.Display
		}
		if strings.HasPrefix(upper, mapping.Source+" ") {
			return mapping.Display + " " + placeTitle035(upper[len(mapping.Source):])
		}
	}
	return placeTitle035(name)
}
