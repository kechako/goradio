package main

import (
	"errors"
	"fmt"

	"github.com/kechako/goradio/audio"
	"github.com/kechako/goradio/rtlfm"
	cli "github.com/urfave/cli/v2"
)

func playRadioCommand(ctx *cli.Context) error {
	sfreq := ctx.String("freq")
	freq, err := rtlfm.ParseFrequency(sfreq)
	if err != nil {
		return ArgumentError("invalid frequency")
	}
	deviceName := ctx.String("device")
	sampleRate := ctx.Int("sample-rate")

	var device *audio.Device
	if deviceName == "" {
		var err error
		device, err = audio.GetDefaultOutputDevice()
		if err != nil {
			return err
		}
	} else {
		var err error
		device, err = audio.GetDevice(deviceName)
		if err != nil {
			return err
		}
	}
	if sampleRate == 0 {
		sampleRate = device.DefaultSampleRate()
	}

	const channels = 1
	var bufferSamples = sampleRate * 10 / 1000

	stream, err := audio.Open[int16](
		audio.WithOutputDevice(device),
		audio.WithOutputChannels(channels),
		audio.WithSampleRate(sampleRate),
		audio.WithBufferSamples(bufferSamples),
	)
	if err != nil {
		return err
	}
	defer stream.Close()

	if err := stream.Start(); err != nil {
		return err
	}
	defer stream.Stop()

	opts := []rtlfm.Option{
		rtlfm.WithSampleRate(sampleRate),
	}
	if ctx.Bool("edge") {
		opts = append(opts, rtlfm.EnableLowerEdgeTuning())
	}
	if ctx.Bool("dc") {
		opts = append(opts, rtlfm.EnableDCBlockingFilter())
	}
	if ctx.Bool("deemp") {
		opts = append(opts, rtlfm.EnableDeEmphasisFilter())
	}
	if ctx.Bool("direct") {
		opts = append(opts, rtlfm.EnableDirectSampling())
	}
	if ctx.Bool("offset") {
		opts = append(opts, rtlfm.EnableOffsetTuning())
	}

	p, err := rtlfm.Play(ctx.Context, freq, opts...)
	if err != nil {
		return fmt.Errorf("failed to play radio: %w", err)
	}
	defer p.Close()

	r := rtlfm.NewFrameReader(p)

	frame := make([]int16, bufferSamples)
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
		}

		if err := r.Read(frame); err != nil {
			return err
		}

		err = stream.Write(frame)
		if errors.Is(err, audio.ErrOutputOverflowed) {
			// ignore
		} else if err != nil {
			return err
		}
	}
	return nil
}
