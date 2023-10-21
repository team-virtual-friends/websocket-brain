package speech

import (
	"bytes"
	"io"
	"math"

	"github.com/iFaceless/godub"
)

func TransposeSound(audioBytes []byte, octaves float64) ([]byte, error) {
	audioSegment, err := godub.NewAudioSegment(audioBytes)
	if err != nil {
		return nil, err
	}

	newSampleRate := (int)((float64)(audioSegment.FrameRate()) * math.Pow(2.0, octaves))
	hipitchSound, err := audioSegment.ForkWithFrameRate(newSampleRate)
	if err != nil {
		return nil, err
	}
	hipitchSound, err = hipitchSound.ForkWithFrameRate(44100)
	if err != nil {
		return nil, err
	}

	// Create a buffer to write data to
	var buf bytes.Buffer

	// Create an io.Writer from the buffer
	writer := io.MultiWriter(&buf)
	godub.NewExporter(writer).
		WithDstFormat("wav").
		Export(hipitchSound)
	return buf.Bytes(), nil
}
