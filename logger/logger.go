package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/OttoLeung-varadise/logmiddleware/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var logQueue = make(chan model.RequestLog, 10000)

func StartLogWriter(db *gorm.DB) {
	const (
		batchSize     = 100
		flushInterval = 500 * time.Millisecond
	)

	for {
		batch := make([]model.RequestLog, 0, batchSize)
		timer := time.NewTimer(flushInterval)
		done := false

		for len(batch) < batchSize && !done {
			select {
			case reqLog := <-logQueue:
				batch = append(batch, reqLog)
			case <-timer.C:
				done = true
			}
		}
		timer.Stop()

		if len(batch) > 0 {
			if err := writeBatchWithGORM(db, batch); err != nil {
				log.Printf("batch log fails: %v, fails counts: %d", err, len(batch))
			}
		}
	}
}

func RequestLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		reqID := uuid.New().String()
		c.Set("request_id", reqID)
		var (
			fileName    string
			fileSize    int64
			content     []byte
			contentType string
		)
		contentType = c.ContentType()

		switch contentType {
		case "multipart/form-data":
			c.Request.ParseMultipartForm(100 << 20)
			file, handler, err := c.Request.FormFile("file")
			if err == nil && file != nil {
				defer file.Close()
				fileName = handler.Filename
				fileSize = handler.Size

				if fileSize > 0 && fileSize <= 100*1024*1024 {
					content, err = io.ReadAll(file)
					if err != nil {
						content = []byte(fmt.Sprintf("read file error: %v", err))
					}
				} else {
					content = []byte("file too large, skip content")
				}
			}
		case "application/json":
			content, _ = io.ReadAll(c.Request.Body)
		}

		c.Next()

		reqLog := model.RequestLog{
			RequestID:   reqID,
			ServiceName: getServiceName(),
			Method:      c.Request.Method,
			Path:        c.Request.URL.Path,
			QueryString: c.Request.URL.RawQuery,
			StatusCode:  c.Writer.Status(),
			RemoteIP:    c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
			RequestTime: time.Since(start).Seconds(),
			CreatedAt:   time.Now(),
			FileName:    fileName,
			FileSize:    fileSize,
			ContentType: contentType,
		}

		go func(reqLog model.RequestLog, content []byte) {
			if len(content) == 0 {
				select {
				case logQueue <- reqLog:
				default:
				}
				return
			}

			var jsonContent model.JSONRawMessage
			if json.Valid(content) {
				jsonContent = content
			} else {
				errMsg := fmt.Sprintf("file content is not valid JSON: %s", string(content[:min(len(content), 10000)]))
				jsonContent, _ = json.Marshal(map[string]string{"error": errMsg})
			}

			reqLog.ContentJSON = jsonContent
			select {
			case logQueue <- reqLog:
			default:
			}
		}(reqLog, content)
	}
}

func writeBatchWithGORM(db *gorm.DB, logs []model.RequestLog) error {
	return db.CreateInBatches(logs, len(logs)).Error
}

func getServiceName() string {
	return ""
}
