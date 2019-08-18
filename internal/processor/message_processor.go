package processor

import (
	"fmt"
	"github.com/twuillemin/modes-viewer/internal/adsbspy"
	"github.com/twuillemin/modes-viewer/internal/plane"
	"github.com/twuillemin/modes/pkg/bds/adsb"
	bds05Fields "github.com/twuillemin/modes/pkg/bds/bds05/fields"
	bds05Messages "github.com/twuillemin/modes/pkg/bds/bds05/messages"
	bds08Messages "github.com/twuillemin/modes/pkg/bds/bds08/messages"
	bds09Fields "github.com/twuillemin/modes/pkg/bds/bds09/fields"
	bds09Messages "github.com/twuillemin/modes/pkg/bds/bds09/messages"
	bds65Fields "github.com/twuillemin/modes/pkg/bds/bds65/fields"
	bds65Messages "github.com/twuillemin/modes/pkg/bds/bds65/messages"
	adsbReader "github.com/twuillemin/modes/pkg/bds/reader"
	modeSCommon "github.com/twuillemin/modes/pkg/modes/common"
	modeSFields "github.com/twuillemin/modes/pkg/modes/fields"
	modeSMessages "github.com/twuillemin/modes/pkg/modes/messages"
	modeSReader "github.com/twuillemin/modes/pkg/modes/reader"
	"math"
)

// ProcessSingleLine processes a signle line of data coming from ADSBSpy or from a file. The expected format is
// like " *8D4BAB4558AB031C446849B72535;1D5D32D0;0A;32AB;" which is the format exported by ADSBSpy.
//
// Params:
//    - str: the line to process
func ProcessSingleLine(planeChannel chan plane.Plane, str string) {

	fmt.Printf("Reading: %s\n", str)

	// Read the line
	messageADSBSpy, err := adsbspy.ReadLine(str)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Convert to a message if possible
	messageModeS, err := modeSReader.ReadMessage(messageADSBSpy.Message)
	if err != nil {
		fmt.Println(err)
		return
	}

	timestamp := messageADSBSpy.Timestamp

	// Check the CRC and get the Address or the Interrogator Identifier
	address, err := modeSReader.CheckCRC(messageModeS, messageADSBSpy.Message, nil, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// For message 11 (Reply to all call) the address is the address of the caller
	if messageModeS.GetDownLinkFormat() == 11 {
		// Update address
		message11 := messageModeS.(*modeSMessages.MessageDF11)
		address = modeSCommon.ICAOAddress(message11.AddressAnnounced.Address)
	}

	// Get the plane
	viewedPlane := plane.CheckoutPlane(timestamp, address)

	// For message with additional content
	switch messageModeS.GetDownLinkFormat() {
	case 17:
		postProcessMessage17(planeChannel, timestamp, viewedPlane, messageModeS.(*modeSMessages.MessageDF17))
	case 18:
		postProcessMessage18(planeChannel, timestamp, viewedPlane, messageModeS.(*modeSMessages.MessageDF18))
	}
}

func postProcessMessage17(planeChannel chan plane.Plane, timestamp uint32, plane *plane.Plane, messageDF17 *modeSMessages.MessageDF17) {

	processADSBMessage(planeChannel, timestamp, plane, messageDF17.MessageExtendedSquitter)
}

func postProcessMessage18(planeChannel chan plane.Plane, timestamp uint32, plane *plane.Plane, messageDF18 *modeSMessages.MessageDF18) {

	if messageDF18.ControlField == modeSFields.ControlFieldADSB || messageDF18.ControlField == modeSFields.ControlFieldADSBReserved {
		processADSBMessage(planeChannel, timestamp, plane, messageDF18.MessageExtendedSquitter)
	} else if messageDF18.ControlField == modeSFields.ControlFieldTISBFineFormat ||
		messageDF18.ControlField == modeSFields.ControlFieldTISBCoarseFormat ||
		messageDF18.ControlField == modeSFields.ControlFieldTISBReservedManagement ||
		messageDF18.ControlField == modeSFields.ControlFieldTISBRelayADSB {
	}
}

func processADSBMessage(planeChannel chan plane.Plane, timestamp uint32, plane *plane.Plane, data []byte) {

	// Get the content
	messageADSB, detectedADSBLevel, errADSB := adsbReader.ReadADSBMessage(plane.ADSBLevel, plane.NICSupplementA, plane.NICSupplementC, data)
	if errADSB != nil {
		fmt.Println(errADSB)
		return
	}

	// Update the plane ADSBLevel
	plane.ADSBLevel = detectedADSBLevel

	if messageADSB == nil {
		return
	}
	planeUpdated := false

	// If message with position
	if message05, ok := messageADSB.(bds05Messages.MessageBDS05); ok {
		if message05.GetCPRFormat() == bds05Fields.CPRFormatEven {
			plane.EvenCPRLatitude = uint32(message05.GetEncodedLatitude())
			plane.EvenCPRLongitude = uint32(message05.GetEncodedLongitude())
			plane.EvenCPRTimestamp = timestamp
		} else {
			plane.OddCPRLatitude = uint32(message05.GetEncodedLatitude())
			plane.OddCPRLongitude = uint32(message05.GetEncodedLongitude())
			plane.OddCPRTimestamp = timestamp
		}

		plane.Altitude = message05.GetAltitude().AltitudeInFeet

		planeUpdated = true
	}

	// If message with altitude - normal plane
	if message09, ok := messageADSB.(bds09Messages.MessageBDS09); ok {

		if format19, okFormat := message09.(*bds09Messages.Format19GroundSpeedNormal); okFormat {
			if format19.VelocityEWNormal.GetStatus() == bds09Fields.VelocityStatusRegular && format19.VelocityNSNormal.GetStatus() == bds09Fields.VelocityStatusRegular {
				plane.AirSpeed = getHypotenuse(format19.VelocityEWNormal.GetVelocity(), format19.VelocityNSNormal.GetVelocity())
				plane.AirSpeedValid = true
			} else {
				plane.AirSpeed = 0
				plane.AirSpeedValid = false
			}
		}

		if format19, okFormat := message09.(*bds09Messages.Format19GroundSpeedSupersonic); okFormat {
			if format19.VelocityEWSupersonic.GetStatus() == bds09Fields.VelocityStatusRegular && format19.VelocityNSSupersonic.GetStatus() == bds09Fields.VelocityStatusRegular {
				plane.AirSpeed = getHypotenuse(format19.VelocityEWSupersonic.GetVelocity(), format19.VelocityNSSupersonic.GetVelocity())
				plane.AirSpeedValid = true
			} else {
				plane.AirSpeed = 0
				plane.AirSpeedValid = false
			}
		}

		if format19, okFormat := message09.(*bds09Messages.Format19AirSpeedNormal); okFormat {
			if format19.AirspeedNormal.GetStatus() == bds09Fields.VelocityStatusRegular {
				plane.AirSpeed = format19.AirspeedNormal.GetAirspeed()
				plane.AirSpeedValid = true
			} else {
				plane.AirSpeed = 0
				plane.AirSpeedValid = false
			}
		}

		if format19, okFormat := message09.(*bds09Messages.Format19AirSpeedSupersonic); okFormat {
			if format19.AirspeedSupersonic.GetStatus() == bds09Fields.VelocityStatusRegular {
				plane.AirSpeed = format19.AirspeedSupersonic.GetAirspeed()
				plane.AirSpeedValid = true
			} else {
				plane.AirSpeed = 0
				plane.AirSpeedValid = false
			}
		}

		// Vertical rate is always present
		if message09.GetVerticalRate().GetStatus() == bds09Fields.VerticalRateStatusRegular {
			if message09.GetVerticalRateSign() == bds09Fields.VRSUp {
				plane.VerticalRate = message09.GetVerticalRate().GetVerticalRate()
			} else {
				plane.VerticalRate = -message09.GetVerticalRate().GetVerticalRate()
			}
			plane.VerticalRateValid = true
		} else {
			plane.VerticalRate = 0
			plane.VerticalRateValid = false
		}

		planeUpdated = true
	}

	// If message with identification
	if message08, ok := messageADSB.(bds08Messages.MessageBDS08); ok {
		if len(plane.Identification) == 0 {
			plane.Identification = string(message08.GetAircraftIdentification())
			planeUpdated = true
		}
	}

	// If message with operational status
	if messageADSB.GetMessageFormat() == adsb.Format31 {
		if message31v1Airborne, ok31v1Airborne := messageADSB.(*bds65Messages.Format31AirborneV1); ok31v1Airborne {
			plane.NICSupplementA = message31v1Airborne.NICSupplement == bds65Fields.NICAOne
		} else if message31v1Surface, ok31v1Surface := messageADSB.(*bds65Messages.Format31SurfaceV1); ok31v1Surface {
			plane.NICSupplementA = message31v1Surface.NICSupplement == bds65Fields.NICAOne
		} else if message31v2Airborne, ok31v2Airborne := messageADSB.(*bds65Messages.Format31AirborneV2); ok31v2Airborne {
			plane.NICSupplementA = message31v2Airborne.NICSupplementA == bds65Fields.NICAOne
		} else if message31v2Surface, ok31v2Surface := messageADSB.(*bds65Messages.Format31SurfaceV2); ok31v2Surface {
			plane.NICSupplementA = message31v2Airborne.NICSupplementA == bds65Fields.NICAOne
			plane.NICSupplementC = message31v2Surface.SurfaceCapabilityClass.NICSupplementC == bds65Fields.NICCSZero
		}
		planeUpdated = true
	}

	if planeUpdated {
		fmt.Printf("==>%v\n", plane.ToString())
		planeChannel <- *plane
	}
}

func getHypotenuse(x, y int) int {
	return int(math.Floor(math.Sqrt(float64(x*x + y*y))))
}
