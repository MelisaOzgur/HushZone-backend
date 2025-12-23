package speedtest

import (
	"crypto/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func Handle(c *gin.Context) {
	const (
		defaultBytes = int64(8_000_000)
		minBytes     = int64(1_000)
		maxBytes     = int64(50_000_000)
		chunkSize    = 32 * 1024
	)

	n := defaultBytes
	if s := c.Query("bytes"); s != "" {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_bytes"})
			return
		}
		n = v
	}

	if n < minBytes {
		n = minBytes
	}
	if n > maxBytes {
		n = maxBytes
	}

	w := c.Writer
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	flusher.Flush()

	buf := make([]byte, chunkSize)
	var sent int64 = 0

	for sent < n {
		remain := n - sent
		toWrite := int64(len(buf))
		if remain < toWrite {
			toWrite = remain
		}

		_, _ = rand.Read(buf[:toWrite])

		_, err := w.Write(buf[:toWrite])
		if err != nil {
			return
		}

		sent += toWrite
		flusher.Flush()

		time.Sleep(2 * time.Millisecond)
	}
}