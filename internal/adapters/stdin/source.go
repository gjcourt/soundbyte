// Package stdin provides a PCMSource that reads from stdin or a named file/pipe.
package stdin

import (
	"bufio"
	"io"
	"os"

	"soundbyte/internal/ports/outbound"
)

// Source reads raw PCM frames from an io.Reader (stdin or file).
type Source struct {
	r *bufio.Reader
}

var _ outbound.PCMSource = (*Source)(nil)

// NewSource creates a PCMSource backed by stdin.
func NewSource() *Source {
	return &Source{r: bufio.NewReader(os.Stdin)}
}

// NewFileSource creates a PCMSource backed by the file at path.
func NewFileSource(path string) (*Source, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &Source{r: bufio.NewReader(f)}, nil
}

// ReadFrame fills buf with exactly len(buf) bytes of PCM data.
func (s *Source) ReadFrame(buf []byte) error {
	_, err := io.ReadFull(s.r, buf)
	return err
}
