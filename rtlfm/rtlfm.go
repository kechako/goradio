package rtlfm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
)

const defaultCommand = "rtl_fm"

type Frequency int

const (
	KiloHertz Frequency = 1000
	MegaHertz           = 1000 * KiloHertz
)

var errParseFrequency = errors.New("failed to parse frequency")

func ParseFrequency(s string) (Frequency, error) {
	if len(s) == 0 {
		return 0, errParseFrequency
	}

	switch unit := s[len(s)-1]; unit {
	case 'K', 'M':
		s = s[:len(s)-1]
		if len(s) == 0 {
			return 0, errParseFrequency
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, errParseFrequency
		}
		if unit == 'K' {
			return Frequency(f * float64(KiloHertz)), nil
		}

		return Frequency(f * float64(MegaHertz)), nil
	default:
		f, err := strconv.Atoi(s)
		if err != nil {
			return 0, errParseFrequency
		}
		return Frequency(f), nil
	}
}

func (f Frequency) String() string {
	if f < KiloHertz {
		return strconv.Itoa(int(f))
	}

	if f < MegaHertz {
		i := f / KiloHertz
		d := f % KiloHertz
		return fmt.Sprintf("%d.%dK", i, d)
	}

	i := f / MegaHertz
	d := f % MegaHertz
	return fmt.Sprintf("%d.%dM", i, d)
}

type Process struct {
	cmd *exec.Cmd
	rc  io.ReadCloser
}

func Play(ctx context.Context, freq Frequency, opts ...Option) (*Process, error) {
	var options playOptions
	for _, opt := range opts {
		opt.apply(&options)
	}

	path, err := commandPath(&options)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(path, makeArguments(freq, &options)...)
	rc, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		rc.Close()
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	return &Process{
		cmd: cmd,
		rc:  rc,
	}, nil
}

func (p *Process) Close() error {
	p.rc.Close()

	err := p.cmd.Process.Signal(os.Interrupt)
	if err != nil {
		p.cmd.Process.Kill()
	}
	return p.cmd.Wait()
}

func (p *Process) Read(b []byte) (n int, err error) {
	return p.rc.Read(b)
}

func makeArguments(freq Frequency, options *playOptions) []string {
	args := []string{
		"-M", "wbfm",
		"-f", freq.String(),
		"-s", "400k",
	}

	if options.sampleRate > 0 {
		args = append(args, "-r", strconv.Itoa(options.sampleRate))
	}

	if options.enableLowerEdgeTuning {
		args = append(args, "-E", "edge")
	}
	if options.enableDCBlockingFilter {
		args = append(args, "-E", "dc")
	}
	if options.enableDeEmphasisFilter {
		args = append(args, "-E", "deemp")
	}
	if options.enableDirectSampling {
		args = append(args, "-E", "direct")
	}
	if options.enableOffsetTuning {
		args = append(args, "-E", "offset")
	}

	return args
}

func commandPath(options *playOptions) (string, error) {
	if options.commandPath != "" {
		return options.commandPath, nil
	}
	path, err := exec.LookPath(defaultCommand)
	if err != nil {
		return "", fmt.Errorf("failed to get rtl_fm path: %w", err)
	}

	return path, nil
}

type playOptions struct {
	commandPath            string
	sampleRate             int
	enableLowerEdgeTuning  bool
	enableDCBlockingFilter bool
	enableDeEmphasisFilter bool
	enableDirectSampling   bool
	enableOffsetTuning     bool
}

type Option interface {
	apply(opts *playOptions)
}

type optionFunc func(opts *playOptions)

func (f optionFunc) apply(opts *playOptions) {
	f(opts)
}

func WithCommandPath(path string) Option {
	return optionFunc(func(opts *playOptions) {
		opts.commandPath = path
	})
}

func WithSampleRate(sampleRate int) Option {
	return optionFunc(func(opts *playOptions) {
		opts.sampleRate = sampleRate
	})
}

func EnableLowerEdgeTuning() Option {
	return optionFunc(func(opts *playOptions) {
		opts.enableLowerEdgeTuning = true
	})
}

func EnableDCBlockingFilter() Option {
	return optionFunc(func(opts *playOptions) {
		opts.enableDCBlockingFilter = true
	})
}

func EnableDeEmphasisFilter() Option {
	return optionFunc(func(opts *playOptions) {
		opts.enableDeEmphasisFilter = true
	})
}

func EnableDirectSampling() Option {
	return optionFunc(func(opts *playOptions) {
		opts.enableDirectSampling = true
	})
}

func EnableOffsetTuning() Option {
	return optionFunc(func(opts *playOptions) {
		opts.enableOffsetTuning = true
	})
}
