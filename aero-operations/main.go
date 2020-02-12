package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/aerospike/aerospike-client-go"
)

var (
	aeroHosts []*aerospike.Host
	aeroNs    string
	aeroSet   string
)

func main() {
	flag.Var(newHostsListValue("127.0.0.1:3000", &aeroHosts), "aero.hosts", "aero host(s) as comma-separated list")
	flag.StringVar(&aeroNs, "aero.ns", "persistent00", "aero namespace")
	flag.StringVar(&aeroSet, "aero.set", "devices", "aero set")

	flag.Parse()

	client, err := aerospike.NewClientWithPolicyAndHost(nil, aeroHosts...)
	if err != nil {
		log.Fatal(err)
	}
	if err := run(client); err != nil {
		log.Fatal(err)
	}
}

func run(client *aerospike.Client) error {
	device := NewDevice()

	key, err := aerospike.NewKey(aeroNs, aeroSet, fmt.Sprintf("device-key-%s-%s", device.ID, device.Token))
	if err != nil {
		return fmt.Errorf("could not create aero key: %w", err)
	}

	wpolicy := aerospike.NewWritePolicy(0, 0)

	bins, _ := device.MarshalBins()
	err = client.PutBins(wpolicy, key, bins...)
	if err != nil {
		return fmt.Errorf("could not put bins for key %s: %w", key, err)
	}

	listOp1 := aerospike.ListAppendOp("events", "eto1")
	listOp2 := aerospike.ListAppendOp("events", "eto2")

	rec, err := client.Operate(nil, key, listOp1, listOp2, aerospike.GetOpForBin("events"))
	if err != nil {
		return fmt.Errorf("could not perform aero operations for key %s: %w", key, err)
	}

	events, _ := rec.Bins["events"].([]interface{})

	log.Printf("got device %s, events %v\n", rec.Key, events)

	return nil
}

type DeviceState uint8

const (
	StateOK DeviceState = 1 + iota
)

type Events map[string]time.Time

type Device struct {
	ID     string
	Token  string
	State  DeviceState
	Events Events
}

func NewDevice() Device {
	return Device{
		ID:     "the-id",
		Token:  "the-app-token",
		State:  StateOK,
		Events: map[string]time.Time{"eto0": time.Now().UTC()},
	}
}

func (d Device) MarshalBins() ([]*aerospike.Bin, error) {
	bb := make([]*aerospike.Bin, 0, 4)
	bb = append(bb, aerospike.NewBin("id", d.ID))
	bb = append(bb, aerospike.NewBin("token", d.Token))
	bb = append(bb, aerospike.NewBin("state", d.State))

	events := make([]string, 0, len(d.Events))
	for token, _ := range d.Events {
		events = append(events, token)
	}
	if len(events) > 0 {
		bb = append(bb, aerospike.NewBin("events", events))
	}

	return bb, nil
}
