package probe

import (
	"encoding/json"
	"strconv"
)

type ProbeInfo struct {
	Key       string  `json:"key"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      string  `json:"name"`
}

func GetProbeLocationList(probes []Probe) ([]byte, error) {
	var plist []ProbeInfo
	for _, p := range probes {

		probeInfo := ProbeInfo{
			Key:       strconv.Itoa(p.Id),
			Latitude:  p.Geometry.Coordinates[1],
			Longitude: p.Geometry.Coordinates[0],
			Name:      strconv.Itoa(p.Id),
		}
		plist = append(plist, probeInfo)
	}
	return json.Marshal(plist)

}
