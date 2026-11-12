package main

import "encoding/json"

type garageCar051 struct{ Key, Name string }

var garageCatalog051 []garageCar051

func garageCars051() []garageCar051 {
	if garageCatalog051 != nil {
		return garageCatalog051
	}
	garageCatalog051 = []garageCar051{{"", "ORIGINAL"}, {"b050_bmw", "M4 WIDEBODY"}, {"b050_eb110", "EB110 SUPER SPORT"}, {"b050_lp670", "LP670"}}
	var added []garageCar051
	if raw, err := assets.ReadFile("assets/b051_garage.json"); err == nil && json.Unmarshal(raw, &added) == nil {
		garageCatalog051 = append(garageCatalog051, added...)
	}
	garageCatalog051 = append(garageCatalog051, garageCar051{"ferrari", "296 CHALLENGE"}, garageCar051{"porsche", "911 TURBO CLASSIC"}, garageCar051{"rm_car2", "MCLAREN P1 GTR"}, garageCar051{"rm_car3", "TRUENO DRIFT"})
	return garageCatalog051
}
