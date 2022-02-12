package audio

import (
	"errors"
	"fmt"
	"time"

	"github.com/gordonklaus/portaudio"
)

func Initialize() error {
	err := portaudio.Initialize()
	if err != nil {
		return fmt.Errorf("failed to initialize audio: %w", err)
	}
	return nil
}

func Terminate() error {
	err := portaudio.Terminate()
	if err != nil {
		return fmt.Errorf("failed to terminate audio: %w", err)
	}

	return nil
}

type SampleType interface {
	float32 | int32 | portaudio.Int24 | int16 | int8 | uint8
}

type Stream[T SampleType] struct {
	stream    *portaudio.Stream
	ibuf      []T
	obuf      []T
	framePool *framePool[T]
}

func Open[T SampleType](opts ...Option) (*Stream[T], error) {
	var options streamOptions
	for _, opt := range opts {
		opt.apply(&options)
	}

	var inputDevice *portaudio.DeviceInfo
	if options.defaultInput {
		device, err := GetDefaultInputDevice()
		if err != nil {
			return nil, err
		}
		inputDevice = device.info
	} else if options.inputDevice != nil {
		inputDevice = options.inputDevice.info
	}
	var outputDevice *portaudio.DeviceInfo
	if options.defaultOutput {
		device, err := GetDefaultOutputDevice()
		if err != nil {
			return nil, err
		}
		outputDevice = device.info
	} else if options.outputDevice != nil {
		outputDevice = options.outputDevice.info
	}

	// channels
	var inputChannels int
	if inputDevice != nil {
		if inputDevice.MaxInputChannels == 0 {
			return nil, errors.New("input device has no input channels")
		}

		inputChannels = options.inputChannels
		if inputChannels <= 0 || inputChannels > inputDevice.MaxInputChannels {
			inputChannels = inputDevice.MaxInputChannels
		}
	}
	var outputChannels int
	if outputDevice != nil {
		if outputDevice.MaxOutputChannels == 0 {
			return nil, errors.New("output device has no output channels")
		}

		outputChannels = options.outputChannels
		if outputChannels <= 0 || outputChannels > outputDevice.MaxOutputChannels {
			outputChannels = outputDevice.MaxOutputChannels
		}
	}

	// latency
	var inputLatency time.Duration
	if inputDevice != nil {
		if inputDevice.DefaultLowInputLatency == 0 {
			return nil, errors.New("input device has no input latency")
		}

		inputLatency = options.inputLatency
		if inputLatency <= 0 || inputLatency > inputDevice.DefaultLowInputLatency {
			inputLatency = inputDevice.DefaultLowInputLatency
		}
	}
	var outputLatency time.Duration
	if outputDevice != nil {
		if outputDevice.DefaultLowOutputLatency == 0 {
			return nil, errors.New("output device has no output latency")
		}

		outputLatency = options.outputLatency
		if outputLatency <= 0 || outputLatency > outputDevice.DefaultLowOutputLatency {
			outputLatency = outputDevice.DefaultLowOutputLatency
		}
	}

	// sampleRate
	sampleRate := float64(options.sampleRate)
	if sampleRate == 0 {
		if inputDevice != nil {
			sampleRate = inputDevice.DefaultSampleRate
		} else if outputDevice != nil {
			sampleRate = outputDevice.DefaultSampleRate
		}
	}

	// buffer samples
	bufferSamples := options.bufferSamples
	if bufferSamples == 0 {
		latency := 10 * time.Millisecond
		if inputLatency > 0 || outputLatency > 0 {
			if inputLatency > outputLatency {
				latency = inputLatency
			} else {
				latency = outputLatency
			}
		}
		// 10ms
		bufferSamples = int(float64(sampleRate) * latency.Seconds())
	}
	var ibuf, obuf []T
	var pool *framePool[T]
	var args []any
	if inputDevice != nil {
		ibuf = make([]T, inputChannels*bufferSamples)
		args = append(args, ibuf)

		pool = newFramePool[T](len(ibuf))
	}
	if outputDevice != nil {
		obuf = make([]T, outputChannels*bufferSamples)
		args = append(args, obuf)
	}

	stream, err := portaudio.OpenStream(portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   inputDevice,
			Channels: inputChannels,
			Latency:  inputLatency,
		},
		Output: portaudio.StreamDeviceParameters{
			Device:   outputDevice,
			Channels: outputChannels,
			Latency:  outputLatency,
		},
		SampleRate:      sampleRate,
		FramesPerBuffer: bufferSamples,
	}, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio stream: %w", err)
	}

	return &Stream[T]{
		stream:    stream,
		ibuf:      ibuf,
		obuf:      obuf,
		framePool: pool,
	}, nil
}

func (s *Stream[T]) Close() error {
	if err := s.stream.Close(); err != nil {
		return fmt.Errorf("failed to close audio stream: %w", err)
	}
	return nil
}

func (s *Stream[T]) Start() error {
	err := s.stream.Start()
	if err != nil {
		return fmt.Errorf("failed to start audio stream: %w", err)
	}
	return nil
}

func (s *Stream[T]) Stop() error {
	err := s.stream.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop audio stream: %w", err)
	}
	return nil
}

var (
	ErrInputOverflowed  = errors.New("input overflowed")
	ErrOutputOverflowed = errors.New("output overflowed")
)

func (s *Stream[T]) Read() (*Frame[T], error) {
	err := s.stream.Read()
	if errors.Is(err, portaudio.InputOverflowed) {
		return nil, ErrInputOverflowed
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read input stream: %w", err)
	}

	frame := s.framePool.Get()
	copy(frame.data, s.ibuf)

	return frame, nil
}

func (s *Stream[T]) Write(frame []T) error {
	if len(frame) != len(s.obuf) {
		return errors.New("invalid frame size")
	}

	copy(s.obuf, frame)

	err := s.stream.Write()
	if errors.Is(err, portaudio.OutputUnderflowed) {
		return ErrOutputOverflowed
	}
	if err != nil {
		return fmt.Errorf("failed to write output stream: %w", err)
	}

	return nil
}

type streamOptions struct {
	defaultInput   bool
	defaultOutput  bool
	inputDevice    *Device
	outputDevice   *Device
	inputChannels  int
	outputChannels int
	inputLatency   time.Duration
	outputLatency  time.Duration
	sampleRate     int
	bufferSamples  int
}

type Option interface {
	apply(opts *streamOptions)
}

type optionFunc func(opts *streamOptions)

func (f optionFunc) apply(opts *streamOptions) {
	f(opts)
}

func WithDefaultInputDevice() Option {
	return optionFunc(func(opts *streamOptions) {
		opts.defaultInput = true
	})
}

func WithDefaultOutputDevice() Option {
	return optionFunc(func(opts *streamOptions) {
		opts.defaultOutput = true
	})
}

func WithInputDevice(device *Device) Option {
	return optionFunc(func(opts *streamOptions) {
		opts.inputDevice = device
	})
}

func WithOutputDevice(device *Device) Option {
	return optionFunc(func(opts *streamOptions) {
		opts.outputDevice = device
	})
}

func WithInputChannels(channels int) Option {
	return optionFunc(func(opts *streamOptions) {
		opts.inputChannels = channels
	})
}

func WithOutputChannels(channels int) Option {
	return optionFunc(func(opts *streamOptions) {
		opts.outputChannels = channels
	})
}

func WithInputLatency(latency time.Duration) Option {
	return optionFunc(func(opts *streamOptions) {
		opts.inputLatency = latency
	})
}

func WithOutputLatency(latency time.Duration) Option {
	return optionFunc(func(opts *streamOptions) {
		opts.outputLatency = latency
	})
}

func WithSampleRate(sampleRate int) Option {
	return optionFunc(func(opts *streamOptions) {
		opts.sampleRate = sampleRate
	})
}

func WithBufferSamples(samples int) Option {
	return optionFunc(func(opts *streamOptions) {
		opts.bufferSamples = samples
	})
}
