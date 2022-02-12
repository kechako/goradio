package main

import (
	"fmt"

	"github.com/kechako/goradio/audio"
	cli "github.com/urfave/cli/v2"
)

func deviceListCommand(ctx *cli.Context) error {
	if ctx.NArg() != 0 {
		return ArgumentError("invalid argument")
	}

	devices, err := getOutputDevices()
	if err != nil {
		return err
	}

	for _, device := range devices {
		fmt.Printf("%s [channels: %d, sample rate: %d]\n",
			device.Name(),
			device.MaxOutputChannels(),
			device.DefaultSampleRate(),
		)
	}

	return nil
}

func getOutputDevices() ([]*audio.Device, error) {
	devices, err := audio.GetDevices()
	if err != nil {
		return nil, err
	}

	var outDevices []*audio.Device
	for _, device := range devices {
		if device.MaxOutputChannels() == 0 {
			continue
		}

		outDevices = append(outDevices, device)
	}

	return outDevices, nil
}

func deviceShowCommand(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return ArgumentError("device name is not specified")
	}

	name := ctx.Args().Get(0)
	device, err := getOutputDevice(name)
	if err != nil {
		return err
	}

	fmt.Println(device.Name())
	fmt.Printf("Channels            : %d\n", device.MaxOutputChannels())
	fmt.Printf("Default sample rate : %d\n", device.DefaultSampleRate())
	fmt.Printf("Default low latency : %s\n", device.DefaultLowOutputLatency())
	fmt.Printf("Default high latency: %s\n", device.DefaultHighOutputLatency())

	return nil
}

func getOutputDevice(name string) (*audio.Device, error) {
	device, err := audio.GetDevice(name)
	if err != nil {
		return nil, err
	}
	if device.MaxOutputChannels() == 0 {
		return nil, audio.ErrDeviceNotFound
	}
	return device, nil
}
