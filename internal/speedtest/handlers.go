package speedtest

import (
	"crypto/rand"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	defaultBytes = int64(8_000_000)
	minBytes     = int64(1_000)
	maxBytes     = int64(120_000_000) 
	chunkSize    = 32 * 1024
)

func clampBytesFromQuery(c *gin.Context) int64 {
	n := defaultBytes
	if s := c.Query("bytes"); s != "" {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return -1
		}
		n = v
	}
	if n < minBytes {
		n = minBytes
	}
	if n > maxBytes {
		n = maxBytes
	}
	return n
}

func HandleDownload(c *gin.Context) {
	n := clampBytesFromQuery(c)
	if n < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_bytes"})
		return
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

		if _, err := w.Write(buf[:toWrite]); err != nil {
			return
		}

		sent += toWrite
		flusher.Flush()

		// küçük throttle (opsiyonel)
		time.Sleep(2 * time.Millisecond)
	}
}

func HandleUpload(c *gin.Context) {
	n := clampBytesFromQuery(c)
	if n < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_bytes"})
		return
	}

	_, _ = io.CopyN(io.Discard, c.Request.Body, n)
	_ = c.Request.Body.Close()

	c.Status(http.StatusNoContent)
}