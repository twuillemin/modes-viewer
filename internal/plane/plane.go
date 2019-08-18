package plane

import (
	"fmt"
	"github.com/twuillemin/modes/pkg/bds/adsb"
	"github.com/twuillemin/modes/pkg/geo"
	"github.com/twuillemin/modes/pkg/modes/common"
	"strings"
)

// Plane is the structure keeping track of the current status of a plane
type Plane struct {
	ICAOAddress        common.ICAOAddress `json:"address"`
	ADSBLevel          adsb.ReaderLevel   `json:"-"`
	Altitude           int                `json:"altitude"`
	Identification     string             `json:"identification"`
	FirstSeenTimestamp uint32             `json:"-"`
	LastSeenTimestamp  uint32             `json:"-"`
	EvenCPRLatitude    uint32             `json:"-"`
	EvenCPRLongitude   uint32             `json:"-"`
	EvenCPRTimestamp   uint32             `json:"-"`
	OddCPRLatitude     uint32             `json:"-"`
	OddCPRLongitude    uint32             `json:"-"`
	OddCPRTimestamp    uint32             `json:"-"`
	AirSpeed           int                `json:"air_speed"`
	AirSpeedValid      bool               `json:"air_speed_valid"`
	VerticalRate       int                `json:"vertical_rate"`
	VerticalRateValid  bool               `json:"vertical_rate_valid"`
	NICSupplementA     bool               `json:"-"`
	NICSupplementC     bool               `json:"-"`
	Latitude           float64            `json:"latitude"`
	Longitude          float64            `json:"longitude"`
}

// ToString returns a very simple representation of the plane
func (plane *Plane) ToString() string {

	lines := make([]string, 0)
	lines = append(lines, fmt.Sprintf("Plane: %v", plane.ICAOAddress.ToString()))
	lines = append(lines, fmt.Sprintf("ADSB Level: %v", plane.ADSBLevel.ToString()))

	if len(plane.Identification) > 0 {
		lines = append(lines, fmt.Sprintf("Flight Identification: %v", plane.Identification))
	}

	plane.Latitude = 0.0
	plane.Longitude = 0.0

	if (plane.EvenCPRTimestamp > 0) && (plane.OddCPRTimestamp > 0) {
		lat, long, err := geo.GetCPRExactPosition(
			plane.EvenCPRLatitude,
			plane.EvenCPRLongitude,
			plane.OddCPRLatitude,
			plane.OddCPRLongitude,
			plane.EvenCPRTimestamp > plane.OddCPRTimestamp)
		if err == nil {

			plane.Latitude = lat
			plane.Longitude = long

			lines = append(lines, fmt.Sprintf("Latitude: %v", lat))
			lines = append(lines, fmt.Sprintf("Longitude: %v", long))

			if (referenceLatitude != 0) && (referenceLongitude != 0) {
				groundDistance := geo.ComputeGroundDistance(referenceLatitude, referenceLongitude, lat, long)
				lines = append(lines, fmt.Sprintf("Ground distance: %v", groundDistance))
			}
		}
	}

	if plane.Altitude > 0 {
		lines = append(lines, fmt.Sprintf("Altitude: %v feet", plane.Altitude))
	}

	if plane.AirSpeedValid {
		lines = append(lines, fmt.Sprintf("Air speed: %v knot", plane.AirSpeed))
	}

	if plane.VerticalRateValid {
		lines = append(lines, fmt.Sprintf("Vertical rate: %v ft/min", plane.VerticalRate))
	}

	return strings.Join(lines, ", ")
}

var referenceLatitude float64
var referenceLongitude float64

// SetReferenceLatitudeLongitude defines the position that can be used to determine the distance
//
// params:
//    - latitude: the reference latitude
//    - longitude: the reference longitude
func SetReferenceLatitudeLongitude(latitude float64, longitude float64) {
	referenceLatitude = latitude
	referenceLongitude = longitude
}
