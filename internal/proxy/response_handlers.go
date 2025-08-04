package proxy

import (
	// "bytes"
	// "encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (ps *ProxyServer) handleStreamingResponse(c *gin.Context, resp *http.Response) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		logrus.Error("Streaming unsupported by the writer, falling back to normal response")
		ps.handleNormalResponse(c, resp)
		return
	}

	buf := make([]byte, 4*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			// logrus.Infof("--- Stream Chunk ---")
			// logrus.Infof("Data: %s", string(buf[:n]))
			// logrus.Infof("--------------------")
			if _, writeErr := c.Writer.Write(buf[:n]); writeErr != nil {
				logUpstreamError("writing stream to client", writeErr)
				return
			}
			flusher.Flush()
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			logUpstreamError("reading from upstream", err)
			return
		}
	}
}

func (ps *ProxyServer) handleNormalResponse(c *gin.Context, resp *http.Response) {
	// bodyBytes, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	logUpstreamError("reading response body", err)
	// 	return
	// }
	// resp.Body.Close()
	// resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // reset body

	// logrus.Infof("--- Proxy Response ---")
	// logrus.Infof("Status Code: %d", resp.StatusCode)
	// headers, _ := json.MarshalIndent(resp.Header, "", "  ")
	// logrus.Infof("Headers: %s", string(headers))
	// logrus.Infof("Body: %s", string(bodyBytes))
	// logrus.Infof("--------------------")

	if _, err := io.Copy(c.Writer, resp.Body); err != nil {
		logUpstreamError("copying response body", err)
	}
}
