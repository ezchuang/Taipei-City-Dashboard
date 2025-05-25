package middleware

import (
	"TaipeiCityDashboardBE/app/elk"
	"TaipeiCityDashboardBE/logs"
	"bytes"
	"github.com/gin-gonic/gin"
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
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		level := logLevelHandler(c)

		if level < levelThreshold {
			return
		}

		msg := elk.ElkMessage{
			Time:     time.Now(),
			LogLevel: logLevelHandler(c),
			ApiRoute: c.Request.RequestURI,
			Server:   "Taipei-City-Dashboard-BE",
			SourceIP: c.ClientIP(),
			Message:  "",
		}
		c.Next()
		msg.Latency = time.Since(msg.Time)

		if level >= logs.InfoLevel { //TODO need CHANGE TO WARN LEVEL
			msg.Message = blw.body.String()
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
