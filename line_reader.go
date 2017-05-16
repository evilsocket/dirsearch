package dirsearch

import (
	"bufio"
	"os"
)

// LineReader will accept the name of a file and offset as argument
// and will return a channel from which lines can be read
// one at a time.
func LineReader(filename string, noff int64) (chan string, error) {
	fp, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	// if offset defined then start from there
	if noff > 0 {
		// and go to the start of the line
		b := make([]byte, 1)
		for b[0] != '\n' {
			noff--
			fp.Seek(noff, os.SEEK_SET)
			fp.Read(b)
		}
		noff++
	}

	out := make(chan string)
	go func() {
		defer fp.Close()
		// we need to close the out channel in order
		// to signal the end-of-data condition
		defer close(out)
		scanner := bufio.NewScanner(fp)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			noff, _ = fp.Seek(0, os.SEEK_CUR)
			out <- scanner.Text()
		}
	}()

	return out, nil
}
