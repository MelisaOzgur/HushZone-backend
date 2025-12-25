package speedtest

import (
	"crypto/rand"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Handle(c *gin.Context) {
	const (
		defaultBytes = int64(25_000_000) 
		minBytes     = int64(1_000_000)  
		maxBytes     = int64(120_000_000) 
		chunkSize    = 256 * 1024        
	)

	n := defaultBytes
	if s := c.Query("bytes"); s != "" {
		if v, err := strconv.ParseInt(s, 10, 64); err == nil {
			n = v
		}
	}

	if n < minBytes {
		n = minBytes
	}
	if n > maxBytes {
		n = maxBytes
	}

	w := c.Writer
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	buf := make([]byte, chunkSize)
	var sent int64 = 0

	for sent < n {
		remain := n - sent
		toWrite := int64(len(buf))
		if remain < toWrite {
			toWrite = remain
		}

		_, _ = rand.Read(buf[:toWrite])

		if _, err := w.Write(buf[:toWrite]); err != nil {
			return
		}

		sent += toWrite
		flusher.Flush()
	}
}