package fileshash

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path"
	"testing"
	"time"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

const (
	FILE_HASH    = "0x15c45aba675b7c49f5def32b4f24e827d478e5dfd712613fd6a5df69a2793b60"
	FILE_CONTENT = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. 
			Nunc finibus ultricies interdum. Nulla porttitor arcu a tincidunt mollis. Aliquam erat volutpat. 
			Maecenas eget ligula mi. Maecenas in ligula non elit fringilla consequat. 
			Morbi non imperdiet odio. Integer viverra ligula a varius tempor. 
			Duis ac velit vel augue faucibus tincidunt ut ac nisl. Nulla sed magna est. 
			Etiam quis nunc in elit ultricies pulvinar sed at felis. 
			Suspendisse fringilla lectus vel est hendrerit pulvinar. 
			Vivamus nec lorem pharetra ligula pulvinar blandit in quis nunc. 
			Cras id eros fermentum mauris tristique faucibus. 
			Praesent vehicula lectus nec ipsum sollicitudin tempus. Nullam et massa velit.`
	PIECE_SIZE = 100
)

type dropableFile struct {
	io.ReadWriteSeeker
	io.Closer
	path string
}

func (f dropableFile) Drop() error {
	return os.Remove(f.path)
}

func TestFileHash_ReadWrite(t *testing.T) {
	require := require.New(t)
	f, err := os.OpenFile("/tmp/testnet.g", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	defer os.Remove(f.Name())
	require.NoError(err)
	tmpDirPath := "/tmp/"
	writer := WrapWriter(f, PIECE_SIZE, 200, func() TmpWriter {
		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)
		tmpPath := path.Join(tmpDirPath, fmt.Sprintf("genesis-tmp-%d", r1.Intn(10000)))
		_ = os.MkdirAll(tmpDirPath, os.ModePerm)
		tmpFh, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
		if err != nil {
			log.Crit("File opening error", "path", tmpPath, "err", err)
		}
		return dropableFile{
			ReadWriteSeeker: tmpFh,
			Closer:          tmpFh,
			path:            tmpPath,
		}
	})

	// write out the (secure) self-hashed file properly
	_, err = writer.Write([]byte(FILE_CONTENT))
	require.NoError(err)
	root, err := writer.Flush()
	require.NoError(err)
	require.Equal(root.Hex(), FILE_HASH)
	f.Close()

	// normal case: correct root hash and content
	{
		f, err = os.OpenFile("/tmp/testnet.g", os.O_RDONLY, 0600)
		require.NoError(err)
		reader := WrapReader(f, 2048, root)
		data := make([]byte, 5)
		n, err := reader.Read(data)
		require.Equal("Lorem", string(data[:]))
		require.NoError(err)
		require.Equal(n, len(data))
		reader.Close()
	}

	// passing the wrong root hash to reader
	{
		f, err = os.OpenFile("/tmp/testnet.g", os.O_RDONLY, 0600)
		require.NoError(err)
		maliciousReader := WrapReader(f, 2048, hash.HexToHash("0x00"))
		data := make([]byte, PIECE_SIZE)
		_, err = maliciousReader.Read(data)
		require.ErrorIs(err, ErrRootMismatch)
		maliciousReader.Close()
	}

	// modify piece data to make the mismatch piece hash
	{
		f, err = os.OpenFile("/tmp/testnet.g", os.O_WRONLY, 0600)
		require.NoError(err)
		f.Seek(300, 0)
		s := []byte("0000000000")
		f.Write(s)
		f.Close()

		f, err = os.OpenFile("/tmp/testnet.g", os.O_RDONLY, 0600)
		maliciousReader := WrapReader(f, 2048, root)
		data := make([]byte, PIECE_SIZE)
		_, err = maliciousReader.Read(data)
		require.ErrorIs(err, ErrHashMismatch)
		maliciousReader.Close()
	}

	// modify a piece hash in file to make the wrong one
	{
		f, err = os.OpenFile("/tmp/testnet.g", os.O_WRONLY, 0600)
		require.NoError(err)
		f.Seek(16, 0)
		s := []byte("0000000000")
		f.Write(s)
		f.Close()

		f, err = os.OpenFile("/tmp/testnet.g", os.O_RDONLY, 0600)
		maliciousReader := WrapReader(f, 2048, root)
		data := make([]byte, PIECE_SIZE)
		_, err = maliciousReader.Read(data)
		require.ErrorIs(err, ErrRootMismatch)
		maliciousReader.Close()
	}
}
