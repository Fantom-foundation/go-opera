package fileshash

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/Fantom-foundation/lachesis-base/hash"

	"github.com/Fantom-foundation/go-opera/utils/ioread"
)

var (
	ErrRootNotFound = errors.New("hashes root not found")
	ErrRootMismatch = errors.New("hashes root mismatch")
	ErrHashMismatch = errors.New("hash mismatch")
	ErrTooMuchMem   = errors.New("hashed file requires too much memory")
	ErrInit         = errors.New("failed to init hashfile")
	ErrPieceRead    = errors.New("failed to read piece")
	ErrClosed       = errors.New("closed")
)

type Reader struct {
	backend io.Reader

	size uint64
	pos  uint64

	pieceSize       uint64
	currentPiecePos uint64
	currentPiece    []byte

	root   hash.Hash
	hashes hash.Hashes

	maxMemUsage uint64

	err error
}

func WrapReader(backend io.Reader, maxMemUsage uint64, root hash.Hash) *Reader {
	return &Reader{
		backend:         backend,
		pos:             0,
		maxMemUsage:     maxMemUsage,
		currentPiecePos: math.MaxUint64,
		root:            root,
	}
}

func (r *Reader) readHashes(n uint64) (hash.Hashes, error) {
	hashes := make(hash.Hashes, n)
	for i := uint64(0); i < n; i++ {
		err := ioread.ReadAll(r.backend, hashes[i][:])
		if err != nil {
			return nil, err
		}
	}
	return hashes, nil
}

func calcHash(piece []byte) hash.Hash {
	hasher := sha256.New()
	hasher.Write(piece)
	return hash.BytesToHash(hasher.Sum(nil))
}

func calcHashesRoot(hashes hash.Hashes, pieceSize, size uint64) hash.Hash {
	hasher := sha256.New()
	hasher.Write(bigendian.Uint32ToBytes(uint32(pieceSize)))
	hasher.Write(bigendian.Uint64ToBytes(size))
	for _, h := range hashes {
		hasher.Write(h.Bytes())
	}
	return hash.BytesToHash(hasher.Sum(nil))
}

func getPiecesNum(size, pieceSize uint64) uint64 {
	if size%pieceSize != 0 {
		return size/pieceSize + 1
	}
	return size / pieceSize
}

func (r *Reader) getPiecesNum(size uint64) uint64 {
	return getPiecesNum(size, r.pieceSize)
}

func (r *Reader) getPiecePos(pos uint64) uint64 {
	return pos / r.pieceSize
}

func (r *Reader) readNewPiece() error {
	// previous piece must be fully read at this point
	// ensure currentPiece has correct size
	maxToRead := r.size - r.pos
	if maxToRead > r.pieceSize {
		maxToRead = r.pieceSize
	}
	if uint64(len(r.currentPiece)) > maxToRead {
		r.currentPiece = r.currentPiece[:maxToRead]
	}
	if uint64(len(r.currentPiece)) < maxToRead {
		r.currentPiece = make([]byte, maxToRead)
	}
	// read currentPiece
	err := ioread.ReadAll(r.backend, r.currentPiece)
	if err != nil {
		return err
	}
	// verify piece hash and advance currentPiecePos
	currentPiecePos := r.getPiecePos(r.pos)
	if calcHash(r.currentPiece) != r.hashes[currentPiecePos] {
		return ErrHashMismatch
	}
	r.currentPiecePos = currentPiecePos
	return nil
}

func (r *Reader) readFromPiece(p []byte) (n int, err error) {
	if r.currentPiecePos != r.getPiecePos(r.pos) {
		// switch to new piece
		err := r.readNewPiece()
		if err != nil {
			return 0, fmt.Errorf("%v: %v", ErrPieceRead, err)
		}
	}
	maxToRead := uint64(len(r.currentPiece))
	if maxToRead > uint64(len(p)) {
		maxToRead = uint64(len(p))
	}
	posInPiece := r.pos % r.pieceSize
	consumed := copy(p[:maxToRead], r.currentPiece[posInPiece:])
	r.pos += uint64(consumed)
	return consumed, nil
}

func memUsageOf(pieceSize, hashesNum uint64) uint64 {
	if hashesNum > math.MaxUint32 {
		return math.MaxUint64
	}
	return pieceSize + hashesNum*128
}

func (r *Reader) init() error {
	buf := make([]byte, 8)
	// read piece size
	err := ioread.ReadAll(r.backend, buf[:4])
	if err != nil {
		return err
	}
	r.pieceSize = uint64(bigendian.BytesToUint32(buf[:4]))
	// read content size
	err = ioread.ReadAll(r.backend, buf)
	if err != nil {
		return err
	}
	r.size = bigendian.BytesToUint64(buf)

	hashesNum := r.getPiecesNum(r.size)
	if memUsageOf(r.pieceSize, hashesNum) > uint64(r.maxMemUsage) {
		return ErrTooMuchMem
	}
	// read piece hashes
	hashes, err := r.readHashes(hashesNum)
	if err != nil {
		return err
	}
	if calcHashesRoot(hashes, r.pieceSize, r.size) != r.root {
		return ErrRootMismatch
	}
	r.hashes = hashes
	return nil
}

func (r *Reader) read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.hashes == nil {
		err := r.init()
		if err != nil {
			return 0, fmt.Errorf("%v: %v", ErrInit, err)
		}
	}
	if r.pos >= r.size {
		return 0, io.EOF
	}
	return r.readFromPiece(p)
}

func (r *Reader) Read(p []byte) (n int, err error) {
	if r.err != nil {
		return 0, r.err
	}
	n, err = r.read(p)
	if err != nil {
		r.err = err
	}
	return n, err
}

func (r *Reader) Close() error {
	r.hashes = nil
	r.err = ErrClosed
	return r.backend.(io.Closer).Close()
}
