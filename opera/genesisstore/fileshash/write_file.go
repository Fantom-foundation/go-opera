package fileshash

import (
	"crypto/sha256"
	"errors"
	hasher "hash"
	"io"

	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/Fantom-foundation/lachesis-base/hash"

	"github.com/Fantom-foundation/go-opera/utils/ioread"
)

type TmpWriter interface {
	io.ReadWriteSeeker
	io.Closer
	Drop() error
}

type tmpWriter struct {
	TmpWriter
	h hasher.Hash
}

type Writer struct {
	backend io.Writer

	openTmp    func(int) TmpWriter
	tmps       []tmpWriter
	tmpReadPos uint64

	size uint64

	pieceSize uint64
}

func WrapWriter(backend io.Writer, pieceSize uint64, openTmp func(int) TmpWriter) *Writer {
	return &Writer{
		backend:   backend,
		openTmp:   openTmp,
		pieceSize: pieceSize,
	}
}

func (w *Writer) writeIntoTmp(p []byte) error {
	if len(p) == 0 {
		return nil
	}
	if w.size/w.pieceSize >= uint64(len(w.tmps)) {
		tmpI := len(w.tmps)
		f := w.openTmp(len(w.tmps))
		if tmpI > 0 {
			err := w.tmps[tmpI-1].Close()
			if err != nil {
				return err
			}
			w.tmps[tmpI-1].TmpWriter = nil
		}
		w.tmps = append(w.tmps, tmpWriter{
			TmpWriter: f,
			h:         sha256.New(),
		})
	}
	currentPosInTmp := w.size % w.pieceSize
	maxToWrite := w.pieceSize - currentPosInTmp
	if maxToWrite > uint64(len(p)) {
		maxToWrite = uint64(len(p))
	}
	n, err := w.tmps[len(w.tmps)-1].Write(p[:maxToWrite])
	w.tmps[len(w.tmps)-1].h.Write(p[:maxToWrite])
	w.size += uint64(n)
	if err != nil {
		return err
	}
	return w.writeIntoTmp(p[maxToWrite:])
}

func (w *Writer) resetTmpReads() error {
	for _, tmp := range w.tmps {
		if tmp.TmpWriter != nil {
			_, err := tmp.Seek(0, io.SeekStart)
			if err != nil {
				return err
			}
		}
	}
	w.tmpReadPos = 0
	return nil
}

func (w *Writer) readFromTmp(p []byte, destructive bool) error {
	if len(p) == 0 {
		return nil
	}
	tmpI := w.tmpReadPos / w.pieceSize
	if tmpI > uint64(len(w.tmps)) {
		return errors.New("all tmp files are consumed")
	}
	if w.tmps[tmpI].TmpWriter == nil {
		w.tmps[tmpI].TmpWriter = w.openTmp(int(tmpI))
	}
	currentPosInTmp := w.tmpReadPos % w.pieceSize
	maxToRead := w.pieceSize - currentPosInTmp
	if maxToRead > uint64(len(p)) {
		maxToRead = uint64(len(p))
	}
	err := ioread.ReadAll(w.tmps[tmpI], p[:maxToRead])
	if err != nil {
		return err
	}
	w.tmpReadPos += maxToRead
	if w.tmpReadPos%w.pieceSize == 0 {
		_ = w.tmps[tmpI].Close()
		if destructive {
			_ = w.tmps[tmpI].Drop()
		}
		w.tmps[tmpI].TmpWriter = nil
	}
	return w.readFromTmp(p[maxToRead:], destructive)
}

func (w *Writer) Write(p []byte) (n int, err error) {
	oldSize := w.size
	err = w.writeIntoTmp(p)
	n = int(w.size - oldSize)
	return
}

func (w *Writer) readFromTmpPieceByPiece(destructive bool, fn func([]byte) error) error {
	err := w.resetTmpReads()
	if err != nil {
		return err
	}
	piece := make([]byte, w.pieceSize)
	for pos := uint64(0); pos < w.size; pos += w.pieceSize {
		end := pos + w.pieceSize
		if end > w.size {
			end = w.size
		}
		err := w.readFromTmp(piece[:end-pos], destructive)
		if err != nil {
			return err
		}
		err = fn(piece[:end-pos])
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) Root() hash.Hash {
	hashes := hash.Hashes{}
	for _, tmp := range w.tmps {
		h := hash.BytesToHash(tmp.h.Sum(nil))
		hashes = append(hashes, h)
	}
	return calcHashesRoot(hashes, w.pieceSize, w.size)
}

func (w *Writer) Flush() (hash.Hash, error) {
	// write piece
	_, err := w.backend.Write(bigendian.Uint32ToBytes(uint32(w.pieceSize)))
	if err != nil {
		return hash.Hash{}, err
	}
	// write size
	_, err = w.backend.Write(bigendian.Uint64ToBytes(w.size))
	if err != nil {
		return hash.Hash{}, err
	}
	// write piece hashes
	hashes := hash.Hashes{}
	for _, tmp := range w.tmps {
		h := hash.BytesToHash(tmp.h.Sum(nil))
		hashes = append(hashes, h)
		_, err = w.backend.Write(h[:])
		if err != nil {
			return hash.Hash{}, err
		}
	}
	root := calcHashesRoot(hashes, w.pieceSize, w.size)
	// write data and drop tmp files
	return root, w.readFromTmpPieceByPiece(true, func(piece []byte) error {
		_, err := w.backend.Write(piece)
		return err
	})
}
