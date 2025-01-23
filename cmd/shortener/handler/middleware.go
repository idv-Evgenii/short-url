package handler

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func LoggerMiddleware() gin.HandlerFunc {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		contentLength := c.Writer.Size()

		logger.WithFields(logrus.Fields{
			"method":        c.Request.Method,
			"uri":           c.Request.RequestURI,
			"status":        statusCode,
			"contentLength": contentLength,
			"duration":      duration.String(),
		}).Info("Handled request")
	}
}

func DecompressGzip() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.String(http.StatusBadRequest, "Failed to decompress request: %v", err)
				c.Abort()
				return
			}
			defer gz.Close()

			body, err := io.ReadAll(gz)
			if err != nil {
				c.String(http.StatusBadRequest, "Failed to read decompressed body: %v", err)
				c.Abort()
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
		}
		c.Next()
	}
}

func CompressGzip() gin.HandlerFunc {
	return func(c *gin.Context) {
		writer := &responseCapture{
			ResponseWriter: c.Writer,
			buf:            bytes.NewBuffer(nil),
		}
		c.Writer = writer

		c.Next()

		if !supportsGzip(c) || (c.Writer.Header().Get("Content-Type") != "application/json" &&
			c.Writer.Header().Get("Content-Type") != "text/html") {
			if _, err := writer.ResponseWriter.Write(writer.buf.Bytes()); err != nil {
				c.String(http.StatusInternalServerError, "Failed to write response")
				return
			}
		}

		var compressedBuf bytes.Buffer
		gz := gzip.NewWriter(&compressedBuf)
		_, err := gz.Write(writer.buf.Bytes())
		gz.Close()
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to compress response")
			return
		}
		c.Writer.Header().Set("Content-Encoding", "gzip")
		c.Writer.Header().Set("Content-Length", fmt.Sprint(compressedBuf.Len()))
		c.Writer.WriteHeader(c.Writer.Status())
		if _, err := c.Writer.Write(compressedBuf.Bytes()); err != nil {
			c.String(http.StatusInternalServerError, "Failed to write compressed response")
			return
		}
	}
}

type responseCapture struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

func (r *responseCapture) Write(data []byte) (int, error) {
	return r.buf.Write(data)
}

func supportsGzip(c *gin.Context) bool {
	return c.GetHeader("Accept-Encoding") != "" && c.GetHeader("Accept-Encoding") != "gzip"
}
