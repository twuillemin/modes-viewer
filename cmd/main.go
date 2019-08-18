// This is the name of our package
// Everything with this package name can see everything
// else inside the same package, regardless of the file they are in
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/twuillemin/modes-viewer/internal/event"
	"github.com/twuillemin/modes-viewer/internal/plane"
	"github.com/twuillemin/modes-viewer/internal/processor"
	"github.com/twuillemin/modes/pkg/bds/adsb"
	"golang.org/x/net/websocket"
	"log"
	"net"
	"os"
	"time"
)

func main() {

	// By default, planes are ADSB level 2 compliant (Europe...)
	plane.SetDefaultADSBLevel(adsb.ReaderLevel2)
	plane.SetReferenceLatitudeLongitude(34.670619, 33.029099)

	adsbSpyServer := flag.String("adsb_spy_server", "localhost", "the name of the ADSBSpy server (default: localhost)")
	adsbSpyPort := flag.Int("adsb_spy_port", 47806, "the port of the ADSBSpy server (default: 47806)")
	fileName := flag.String("file", "", "the name of the file to be processed")
	flag.Parse()

	planeEventMux := event.CreatePlaneUpdateMultiplexer()

	// Create an Echo server with the basic middleware
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("events", buildEndpoint(planeEventMux))
	e.Static("/", "web")

	// Start reception from ADSBSpy
	if len(*fileName) > 0 {
		go readFile("example/example.txt", planeEventMux.GetPublishingQueue())
	} else {
		go readFromADSBSpy(*adsbSpyServer, *adsbSpyPort, planeEventMux.GetPublishingQueue())
	}

	err := e.Start("127.0.0.1:8081")
	if err != nil {
		log.Fatalf("Unable to start server due to error: \"%v\"", err)
	}
}

func buildEndpoint(mux event.PlaneUpdateMultiplexer) func(c echo.Context) error {

	handler := func(c echo.Context) error {
		websocket.Handler(func(ws *websocket.Conn) {

			planeChannel, stopChannel := mux.CreateClient()

			defer func() {
				if closeError := ws.Close(); closeError != nil {
					log.Println(closeError)
				}
			}()

			for exit := false; exit == false; {
				select {
				case updatedPlane := <-planeChannel:
					// Convert the plane to JSON
					planeJson, err := json.Marshal(updatedPlane)
					if err != nil {
						c.Logger().Error(err)
					}

					fmt.Println(string(planeJson))

					// Write
					err = websocket.Message.Send(ws, string(planeJson))
					if err != nil {
						c.Logger().Error(err)
					}

				case <-stopChannel:
					exit = true
				}
			}
		}).ServeHTTP(c.Response(), c.Request())

		return nil
	}

	return handler
}

func readFromADSBSpy(
	adsbSpyServer string,
	adsbSpyPort int,
	planeChannel chan plane.Plane) {

	address := fmt.Sprintf("%v:%v", adsbSpyServer, adsbSpyPort)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}

	bReader := bufio.NewReader(conn)
	for {
		if fileLine, readErr := bReader.ReadBytes('\n'); readErr == nil {
			processor.ProcessSingleLine(planeChannel, string(fileLine))
		} else {
			log.Fatal(readErr)
		}
	}
}

func readFile(fileName string, planeChannel chan plane.Plane) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		errClose := file.Close()
		if errClose != nil {
			log.Fatal(errClose)
		}
	}()

	lines := make([]string, 0, 1000)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	lineIndex := 0

	for finished := false; finished == false; {
		select {
		case <-ticker.C:
			processor.ProcessSingleLine(planeChannel, lines[lineIndex])
			lineIndex++
			if lineIndex > len(lines) {
				finished = true
			}
		}
	}
}
