package fileshash

import (
	"errors"
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

type Writer struct {
	backend io.Writer

	openTmp    func() TmpWriter
	tmps       []TmpWriter
	tmpSize    uint64
	tmpReadPos uint64

	size uint64

	pieceSize uint64
}

func WrapWriter(backend io.Writer, pieceSize, tmpFileLen uint64, openTmp func() TmpWriter) *Writer {
	return &Writer{
		backend:   backend,
		openTmp:   openTmp,
		pieceSize: pieceSize,
		tmpSize:   tmpFileLen,
	}
}

func (w *Writer) writeIntoTmp(p []byte) error {
	if len(p) == 0 {
		return nil
	}
	if w.size/w.tmpSize >= uint64(len(w.tmps)) {
		w.tmps = append(w.tmps, w.openTmp())
	}
	currentPosInTmp := w.size % w.tmpSize
	maxToWrite := w.tmpSize - currentPosInTmp
	if maxToWrite > uint64(len(p)) {
		maxToWrite = uint64(len(p))
	}
	n, err := w.tmps[len(w.tmps)-1].Write(p[:maxToWrite])
	w.size += uint64(n)
	if err != nil {
		return err
	}
	return w.writeIntoTmp(p[maxToWrite:])
}

func (w *Writer) resetTmpReads() error {
	for _, tmp := range w.tmps {
		_, err := tmp.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}
	}
	w.tmpReadPos = 0
	return nil
}

func (w *Writer) readFromTmp(p []byte, destructive bool) error {
	if len(p) == 0 {
		return nil
	}
	tmpI := w.tmpReadPos / w.tmpSize
	if tmpI > uint64(len(w.tmps)) {
		return errors.New("all tmp files are consumed")
	}
	currentPosInTmp := w.tmpReadPos % w.tmpSize
	maxToRead := w.tmpSize - currentPosInTmp
	if maxToRead > uint64(len(p)) {
		maxToRead = uint64(len(p))
	}
	err := ioread.ReadAll(w.tmps[tmpI], p[:maxToRead])
	if err != nil {
		return err
	}
	w.tmpReadPos += maxToRead
	if destructive && w.tmpReadPos%w.tmpSize == 0 {
		// erase tmp data piece to avoid double disk usage
		_ = w.tmps[tmpI].Close()
		_ = w.tmps[tmpI].Drop()
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
	err = w.readFromTmpPieceByPiece(false, func(piece []byte) error {
		h := calcHash(piece)
		hashes = append(hashes, h)
		_, err := w.backend.Write(h[:])
		return err
	})
	if err != nil {
		return hash.Hash{}, err
	}
	root := calcHashesRoot(hashes, w.pieceSize, w.size)
	// write data and drop tmp files
	return root, w.readFromTmpPieceByPiece(true, func(piece []byte) error {
		_, err := w.backend.Write(piece)
		return err
	})
}
