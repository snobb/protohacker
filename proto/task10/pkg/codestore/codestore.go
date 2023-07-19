package codestore

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

var (
	// glocal data store.
	store = make(map[string][][]byte)
	mu    sync.Mutex
)

// CodeStore is a VCS structure
type CodeStore struct {
	r    *bufio.Reader
	w    io.Writer
	addr net.Addr
}

// New returns a pointer to the new CodeStore instance
func New(rw io.ReadWriter, addr net.Addr) *CodeStore {
	return &CodeStore{
		r:    bufio.NewReader(rw),
		w:    rw,
		addr: addr,
	}
}

// Handle handles a signle tcp connection
func (c *CodeStore) Handle(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		c.send("READY")

		line, err := c.r.ReadString('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				c.logDebug("error: %s", err.Error())
				c.send("ERR illegal method:")
			}
			return
		}

		c.logDebug("line: %s", strings.TrimSpace(line))

		toks := strings.Split(strings.TrimSpace(line), " ")
		if len(line) == 0 || len(toks) == 0 {
			c.send("ERR illegal method:")
			return
		}

		cmd, args := strings.ToLower(toks[0]), toks[1:]

		switch strings.TrimSpace(cmd) {
		case "list":
			dir, err := c.validateListArgs(args)
			if err != nil {
				c.send(err.Error())
				continue
			}

			c.listFiles(dir)

		case "get":
			fname, rev, err := c.validateGetArgs(args)
			if err != nil {
				c.send(err.Error())
				continue
			}

			c.getFile(fname, rev)

		case "put":
			fname, n, err := c.validatePutArgs(args)
			if err != nil {
				c.send(err.Error())
				continue
			}

			c.putFile(fname, n)

		case "help":
			c.send("OK usage: HELP|GET|PUT|LIST")

		case "clear-data":
			for k := range store {
				delete(store, k)
			}

		default:
			c.send(fmt.Sprintf("ERR illegal method: %s", cmd))
			return
		}
	}
}

func (c *CodeStore) validateGetArgs(args []string) (string, int, error) {
	if len(args) < 1 || len(args) > 2 {
		return "", -1, errors.New("ERR usage: GET file [revision]")
	}

	fname := strings.TrimSpace(args[0])
	if err := c.validateName(fname); err != nil {
		return "", -1, err
	}

	rev := -1
	if len(args) == 2 {
		rr := strings.TrimSpace(args[1])
		if rr[0] == 'r' {
			rr = rr[1:]
		}

		var err error
		rev, err = strconv.Atoi(rr)
		if err != nil {
			return "", -1, errors.New("ERR no such revision")
		}
	}

	return fname, rev, nil
}

func (c *CodeStore) validatePutArgs(args []string) (string, int, error) {
	if len(args) != 2 {
		return "", -1, errors.New("ERR usage: PUT file length newline data")
	}

	fname := strings.TrimSpace(args[0])
	if err := c.validateName(fname); err != nil {
		return "", -1, err
	}

	sz, _ := strconv.Atoi(strings.TrimSpace(args[1]))
	return fname, sz, nil
}

func (c *CodeStore) validateListArgs(args []string) (string, error) {
	if len(args) != 1 {
		return "", errors.New("ERR usage: LIST dir")
	}

	fname := strings.TrimSpace(args[0])
	if err := c.validateName(fname); err != nil {
		return "", err
	}

	return fname, nil
}

func (c *CodeStore) validatePutData(data []byte) error {
	for i := 0; i < len(data); i++ {
		if data[i] > unicode.MaxASCII {
			return errors.New("ERR text files only")
		}
	}

	return nil
}

func (c *CodeStore) validateName(name string) error {
	if len(name) == 0 || name[0] != '/' {
		return errors.New("ERR illegal file name")
	}

	var prev rune
	for _, ch := range name {
		if ch == '/' && prev == '/' {
			return errors.New("ERR illegal file name")
		}

		prev = ch

		if unicode.IsDigit(ch) || unicode.IsLetter(ch) ||
			ch == '.' || ch == '/' || ch == '-' || ch == '_' {
			continue
		}

		return errors.New("ERR illegal file name")
	}

	return nil
}

func (c *CodeStore) getFile(fname string, rev int) {
	data, err := c.getValue(fname, rev)
	if err != nil {
		c.send(err.Error())
		return
	}

	c.send(fmt.Sprintf("OK %d", len(data)))
	if _, err := c.w.Write(data); err != nil {
		c.logDebug("get: write data error: %s", err.Error())
		return
	}
}

func (c *CodeStore) putFile(fname string, n int) {
	var rev int

	if n == 0 {
		rev = c.putValue(fname, []byte{})
		c.send(fmt.Sprintf("OK r%d", rev))
	}

	buf := make([]byte, n)
	if _, err := io.ReadFull(c.r, buf); err != nil {
		c.logDebug("put: readFull error: %s", err.Error())
		return
	}

	if err := c.validatePutData(buf); err != nil {
		c.send(err.Error())
		return
	}

	if lst, ok := store[fname]; ok {
		for i, sdata := range lst {
			if bytes.Equal(sdata, buf) {
				c.send(fmt.Sprintf("OK r%d", i+1))
				return
			}
		}
	}

	rev = c.putValue(fname, buf)
	c.send(fmt.Sprintf("OK r%d", rev))
}

func (c *CodeStore) listFiles(dir string) {
	if dir[len(dir)-1] != '/' {
		dir += "/"
	}

	dirs := make(map[string]struct{})
	files := []struct {
		string
		int
	}{}

	for k, v := range store {
		if strings.HasPrefix(k, dir) {
			value := k[len(dir):]
			idx := strings.IndexRune(value, '/')
			if idx != -1 {
				dirs[value[:idx+1]] = struct{}{}
			} else {
				files = append(files, struct {
					string
					int
				}{
					value,
					len(v),
				})
			}
		}
	}

	c.send(fmt.Sprintf("OK %d", len(dirs)+len(files)))

	dirSlice := make([]string, 0, len(dirs))
	for k := range dirs {
		dirSlice = append(dirSlice, k)
	}

	sort.Strings(dirSlice)
	sort.Slice(files, func(i, j int) bool {
		return files[i].string < files[j].string
	})

	for _, d := range dirSlice {
		c.send(fmt.Sprintf("%s DIR", d))
	}

	for _, f := range files {
		c.send(fmt.Sprintf("%s r%d", f.string, f.int))
	}
}

func (c *CodeStore) send(msg string) {
	c.logDebug(msg)
	if _, err := fmt.Fprintln(c.w, msg); err != nil {
		c.logDebug("send: Fprintf error: %s", err.Error())
	}
}

func (c *CodeStore) getValue(file string, rev int) ([]byte, error) {
	fstore, ok := store[file]
	if !ok {
		return nil, fmt.Errorf("ERR no such file")
	}

	if rev == -1 {
		return fstore[len(fstore)-1], nil
	}

	if rev < 1 || rev > len(fstore) {
		return nil, fmt.Errorf("ERR no such revision")
	}

	return fstore[rev-1], nil
}

func (c *CodeStore) putValue(file string, data []byte) int {
	mu.Lock()
	defer mu.Unlock()

	fstore, ok := store[file]
	if !ok {
		fstore = [][]byte{}
	}

	store[file] = append(fstore, data)
	return len(store[file])
}

func (c *CodeStore) logDebug(msg string, args ...interface{}) {
	log.Printf(fmt.Sprintf("%s: %s", c.addr.String(), msg), args...)
}
