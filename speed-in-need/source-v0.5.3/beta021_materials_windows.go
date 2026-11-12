package main

import "strings"

// Reflection, direct specular and a roughness floor are explicit surface traits.
// Roughness maps may add variation without making dry paving or leaves glossy.
func beta021Surface(m pbrMaterial) (pbrMaterial, pos3) {
	name := strings.ToLower(m.Name)
	response := pos3{.035, .18, .72}
	switch {
	case name == "city window glow":
		m.Metal, m.Rough, m.Coat = .25, .28, .25
		response = pos3{.70, .80, .18}
	case m.Glass:
		m.Metal = 0
		m.Rough = .12
		response = pos3{.85, .80, .09}
	case m.Paint:
		response = pos3{1.0, .95, .16}
	case m.Surface == "foliage" || m.Alpha == "MASK":
		m.Metal = 0
		m.Coat = 0
		response = pos3{0, .045, .94}
	case m.Surface == "rubber" || strings.Contains(strings.ToLower(m.Name), "tire"):
		m.Metal = 0
		m.Coat = 0
		response = pos3{.008, .12, .90}
	case m.Surface == "water":
		m.Metal = 0
		m.Coat = 0
		m.Rough = .26
		response = pos3{.24, .35, .26}
	case m.Surface == "road":
		m.Metal = 0
		m.Coat = 0
		response = pos3{.025, .18, .86}
	case m.Surface == "ground":
		m.Metal = 0
		m.Coat = 0
		response = pos3{0, .08, .94}
	case remasterVehicleLamps.Active && (name == "primary" || name == "secondary" || name == "trueno"):
		m.Metal, m.Rough, m.Coat = .12, .24, .85
		response = pos3{.95, .90, .16}
	case strings.Contains(name, "chrome") || strings.Contains(name, "mirror"):
		m.Metal, m.Rough = .95, .14
		response = pos3{1, 1, .09}
	case strings.Contains(name, "rim") || strings.Contains(name, "rotor") || strings.Contains(name, "exhaust") || strings.Contains(name, "brake_disc") || strings.Contains(name, "cockpit_metal") || name == "bronze" || name == "disk":
		m.Metal, m.Rough = .86, .27
		response = pos3{.95, 1, .16}
	case m.Metal > .55:
		response = pos3{.95, 1, .18}
	}
	return m, response
}
