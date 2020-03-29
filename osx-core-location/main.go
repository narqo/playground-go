package main

import (
	"fmt"
)

func main() {
	loc, err := CurrentLocation()
	fmt.Printf("got location: latitude %f, logitude %f, err %v\n", loc.Coordinate.Latitude, loc.Coordinate.Longitude, err)
}
