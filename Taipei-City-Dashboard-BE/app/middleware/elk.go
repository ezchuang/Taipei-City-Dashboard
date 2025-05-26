package middleware

import (
	"TaipeiCityDashboardBE/app/elk"
	"TaipeiCityDashboardBE/logs"
	"bytes"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func ElkLoggingMiddleware(levelThreshold logs.LogLevel) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		level := logLevelHandler(c)

		if level < levelThreshold {
			return
		}

		var requestBody []byte

		//logging request body
		if level >= logs.InfoLevel { //TODO need CHANGE TO WARN LEVEL

			if c.Request.Body != nil {
				data, err := io.ReadAll(c.Request.Body)
				if err == nil {
					c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
					requestBody = data
				} else {
					logs.Warn("Failed to read request body: %v", err)
				}
			}
		}

		c.Next()
		latency := time.Since(now)
		var responseBody string
		//logging response body
		if level >= logs.InfoLevel { //TODO need CHANGE TO WARN LEVEL
			responseBody = blw.body.String()
		}

		msg := elk.ElkMessage{
			Time:            now,
			LogLevel:        logLevelHandler(c),
			ApiRoute:        c.FullPath(),
			RequestUrl:      c.Request.RequestURI,
			ApplicationName: "Taipei-City-Dashboard-BE",
			SourceIP:        c.ClientIP(),
			Message:         "",
			RequestBody:     string(requestBody),
			ResponseBody:    responseBody,
			Latency:         latency.Microseconds(),
			HttpStatus:      c.Writer.Status(),
		}

		//publish message to ELK queue
		_ = elk.MessageWorker.Submit(msg)
	}
}

func logLevelHandler(c *gin.Context) logs.LogLevel {
	status := c.Writer.Status()
	switch {
	case status >= 500:
		return logs.ErrorLevel
	case status >= 400:
		return logs.WarnLevel
	default:
		return logs.InfoLevel
	}
}
