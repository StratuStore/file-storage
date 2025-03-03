package fileio

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"runtime"
	"strconv"
	"sync"
	"testing"
)

const BufferSize = 512

func prepareReaderFileMock(t interface {
	mock.TestingT
	Cleanup(func())
}) (File, *os.File) {

	fileMock := NewMockFile(t)

	testFile, err := os.CreateTemp("", "FuzzReader_Seek_*.txt")
	require.Nil(t, err)

	mx := sync.RWMutex{}

	fileMock.On("openForReading").Return(testFile, nil)
	fileMock.On("rwMx").Return(&mx)
	fileMock.On("version").Return(1)
	fileMock.On("Closed").Return(false)
	fileMock.On("Reader", mock.Anything).Return(newFileReader(fileMock, BufferSize))

	return fileMock, testFile
}

func TestReader_ReadSequentiallyPerSector_ReturnFullFile(t *testing.T) {
	testData := []struct {
		fileBytes []byte
		sectors   []int // < len(fileBytes)
	}{
		{
			[]byte("hello and welcome"),
			[]int{4, 9, 13},
		},
		{
			[]byte("hq"),
			[]int{1},
		},
	}

	for _, tt := range testData {
		t.Run(string(tt.fileBytes), func(t *testing.T) {
			t.Parallel()
			info := tt.fileBytes
			fileMock, testFile := prepareReaderFileMock(t)
			defer testFile.Close()

			_, err := testFile.Write(info)
			require.Nil(t, err)
			n, err := testFile.Seek(0, 0)
			require.EqualValues(t, 0, n, "position of the file at the beginning must be 0")
			require.Nil(t, err, "position of the file at the beginning must be 0")

			r, err := fileMock.Reader(BufferSize)
			assert.Nil(t, err, "unable to get reader")
			require.NotNil(t, r, "reader must not be nil")

			data, err := io.ReadAll(r)
			assert.Nil(t, err, "ReadAll call must have nil error")
			assert.Equal(t, info, data, "info and data must be equal")

			n, err = r.Seek(0, 0)
			assert.EqualValues(t, 0, n, "position of the file should be 0")
			assert.Nil(t, err, "position of the file should be 0")

			sectors := append(tt.sectors, len(info))
			b := make([]byte, len(info))
			var prevSector int

			for _, sector := range sectors {
				t.Run(t.Name()+"_"+strconv.Itoa(sector), func(t *testing.T) {
					n64, err := r.Read(b[prevSector:sector])
					assert.EqualValues(t, sector-prevSector, n64, "size of read data must be equal to diff between borders", sector, "-", prevSector)
					assert.Equal(t, info[prevSector:sector], b[prevSector:sector], "sector of read data = sector of initial data")
					assert.Nil(t, err, "error must be nil")
					prevSector = sector
					runtime.GC()
				})
			}

			assert.Equal(t, info, b, "full data must be equal")
		})
	}
}

func TestReader_ReadWithSeekFromBeginningPerSector_ReturnFullFile(t *testing.T) {
	testData := []struct {
		fileBytes []byte
		sectors   []int // < len(fileBytes)
		order     []int // = len(sectors) + 1
	}{
		{
			[]byte("hello and welcome"),
			[]int{4, 7, 10, 13},
			[]int{2, 1, 4, 0, 3},
		},
		{
			[]byte("hello and welcome"),
			[]int{4, 9, 10, 13},
			[]int{1, 2, 3, 4, 0},
		},
		{
			[]byte("hq"),
			[]int{1},
			[]int{0, 1},
		},
		{
			[]byte("hq"),
			[]int{1},
			[]int{1, 0},
		},
	}

	for _, tt := range testData {
		t.Run(string(tt.fileBytes)+strconv.Itoa(tt.order[0]), func(t *testing.T) {
			t.Parallel()
			info := tt.fileBytes
			fileMock, testFile := prepareReaderFileMock(t)
			defer testFile.Close()

			_, err := testFile.Write(info)
			require.NoError(t, err)
			n, err := testFile.Seek(0, 0)
			require.EqualValues(t, 0, n, "position of the file at the beginning must be 0")
			require.NoError(t, err, "position of the file at the beginning must be 0")

			r, err := fileMock.Reader(BufferSize)
			assert.NoError(t, err, "unable to get reader")
			require.NotNil(t, r, "reader must not be nil")

			fullSectors := append(tt.sectors, len(info))
			var startOfSector int
			sectors := make([]struct{ start, end int }, len(fullSectors))

			// formSectors
			for num, endOfSector := range fullSectors {
				sectors[tt.order[num]] = struct{ start, end int }{startOfSector, endOfSector}
				startOfSector = endOfSector
			}

			b := make([]byte, len(info))

			for _, sector := range sectors {
				t.Run(t.Name()+"_"+strconv.Itoa(sector.start), func(t *testing.T) {
					n, err := r.Seek(int64(sector.start), 0)
					assert.EqualValues(t, sector.start, n, "absolute position must be equal to start of the sector")
					assert.NoError(t, err, "error must be nil")

					n64, err := r.Read(b[sector.start:sector.end])
					assert.EqualValues(t, sector.end-sector.start, n64, "size of read data must be equal to diff between borders", sector.end, "-", sector.start)
					assert.Equal(t, info[sector.start:sector.end], b[sector.start:sector.end], "sector of read data = sector of initial data")
					assert.NoError(t, err, "error must be nil")
					if sector.start%2 == 0 {
						runtime.GC()
					}
				})
			}

			assert.Equal(t, info, b, "full data must be equal")
		})
	}
}
