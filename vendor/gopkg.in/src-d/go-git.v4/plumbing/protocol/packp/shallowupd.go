package packp

import (
	"bytes"
	"fmt"
	"io"

	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/format/pktline"
)

const (
	shallowLineLen   = 48
	unshallowLineLen = 50
)

type ShallowUpdate struct {
	Shallows   []plumbing.Hash
	Unshallows []plumbing.Hash
}

func (r *ShallowUpdate) Decode(reader io.Reader) error {
	s := pktline.NewScanner(reader)

	for s.Scan() {
		line := s.Bytes()

		var err error
		switch {
		case bytes.HasPrefix(line, shallow):
			err = r.decodeShallowLine(line)
		case bytes.HasPrefix(line, unshallow):
			err = r.decodeUnshallowLine(line)
		case bytes.Compare(line, pktline.Flush) == 0:
			return nil
		}

		if err != nil {
			return err
		}
	}

	return s.Err()
}

func (r *ShallowUpdate) decodeShallowLine(line []byte) error {
	hash, err := r.decodeLine(line, shallow, shallowLineLen)
	if err != nil {
		return err
	}

	r.Shallows = append(r.Shallows, hash)
	return nil
}

func (r *ShallowUpdate) decodeUnshallowLine(line []byte) error {
	hash, err := r.decodeLine(line, unshallow, unshallowLineLen)
	if err != nil {
		return err
	}

	r.Unshallows = append(r.Unshallows, hash)
	return nil
}

func (r *ShallowUpdate) decodeLine(line, prefix []byte, expLen int) (plumbing.Hash, error) {
	if len(line) != expLen {
		return plumbing.ZeroHash, fmt.Errorf("malformed %s%q", prefix, line)
	}

	raw := string(line[expLen-40 : expLen])
	return plumbing.NewHash(raw), nil
}
