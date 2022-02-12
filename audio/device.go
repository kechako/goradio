package audio

import (
	"errors"
	"fmt"
	"time"

	"github.com/gordonklaus/portaudio"
)

type Device struct {
	info *portaudio.DeviceInfo
}

func (d *Device) Name() string                            { return d.info.Name }
func (d *Device) MaxInputChannels() int                   { return d.info.MaxInputChannels }
func (d *Device) MaxOutputChannels() int                  { return d.info.MaxOutputChannels }
func (d *Device) DefaultLowInputLatency() time.Duration   { return d.info.DefaultLowInputLatency }
func (d *Device) DefaultLowOutputLatency() time.Duration  { return d.info.DefaultLowOutputLatency }
func (d *Device) DefaultHighInputLatency() time.Duration  { return d.info.DefaultHighInputLatency }
func (d *Device) DefaultHighOutputLatency() time.Duration { return d.info.DefaultHighOutputLatency }
func (d *Device) DefaultSampleRate() int                  { return int(d.info.DefaultSampleRate) }

func GetDevices() ([]*Device, error) {
	infos, err := portaudio.Devices()
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	if len(infos) == 0 {
		return nil, nil
	}

	devices := make([]*Device, len(infos))
	for i, info := range infos {
		devices[i] = &Device{info: info}
	}

	return devices, nil
}

func GetDefaultInputDevice() (*Device, error) {
	info, err := portaudio.DefaultInputDevice()
	if err != nil {
		return nil, fmt.Errorf("failed to get default input device: %w", err)
	}

	return &Device{info: info}, nil
}

func GetDefaultOutputDevice() (*Device, error) {
	info, err := portaudio.DefaultOutputDevice()
	if err != nil {
		return nil, fmt.Errorf("failed to get default output device: %w", err)
	}

	return &Device{info: info}, nil
}

var ErrDeviceNotFound = errors.New("device not found")

func GetDevice(name string) (*Device, error) {
	devices, err := GetDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}
	for _, device := range devices {
		if device.Name() == name {
			return device, nil
		}
	}

	return nil, ErrDeviceNotFound
}
