package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/OttoLeung-varadise/logmiddleware/model"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// FiberRequestLogMiddleware Fiber 版本的日志中间件
func FiberRequestLogMiddleware(pathFilter []string) fiber.Handler {
	return func(c *fiber.Ctx) error {

		if len(pathFilter) > 0 {
			path := c.Path()
			filterMap := make(map[string]struct{}, len(pathFilter))
			for _, s := range pathFilter {
				filterMap[s] = struct{}{}
				if strings.HasSuffix(s, "/*") {
					prefix := strings.TrimSuffix(s, "/*")
					if strings.HasPrefix(path, prefix) || path == prefix {
						return c.Next() // Fiber 需返回 error
					}
				}
			}
			if _, exists := filterMap[path]; exists {
				return c.Next()
			}
		}

		start := time.Now()
		reqID := uuid.New().String()
		c.Locals("request_id", reqID)

		var (
			fileName    string
			fileSize    int64
			content     []byte
			contentType = c.Get("Content-Type")
		)

		switch {
		case strings.Contains(contentType, "multipart/form-data"):
			// Fiber 处理文件上传
			file, err := c.FormFile("file")
			if err == nil && file != nil {
				fileName = file.Filename
				fileSize = file.Size
				if fileSize > 0 && fileSize <= 100*1024*1024 {
					src, err := file.Open()
					if err != nil {
						content = []byte(fmt.Sprintf("open file error: %v", err))
					} else {
						defer src.Close()
						content, err = io.ReadAll(src)
						if err != nil {
							content = []byte(fmt.Sprintf("read file error: %v", err))
						}
					}
				} else {
					content = []byte("file too large, skip content")
				}
			}

		case strings.Contains(contentType, "application/json"):
			// Fiber 读取 JSON  body（复制一份避免被消费）
			content = c.Body()
			contentStr := string(content)
			c.Set("body", contentStr)
			c.Context().Request.SetBodyStream(bytes.NewReader(content), len(content))
		}

		if err := c.Next(); err != nil {
			return err
		}

		reqLog := model.RequestLog{
			RequestID:   reqID,
			ServiceName: serviceName,
			Method:      c.Method(),
			Path:        c.Path(),
			QueryString: c.Context().QueryArgs().String(),
			StatusCode:  c.Response().StatusCode(),
			RemoteIP:    c.IP(),
			UserAgent:   c.Get("User-Agent"),
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
					log.Println("log queue is full, drop log")
				}
				return
			}

			var jsonContent model.JSONRawMessage
			if json.Valid(content) {
				jsonContent = content
			} else {
				errMsg := fmt.Sprintf("content is not valid JSON: %s", string(content[:min(len(content), 10000)]))
				jsonContent, _ = json.Marshal(map[string]string{"error": errMsg})
			}
			reqLog.ContentJSON = jsonContent

			select {
			case logQueue <- reqLog:
			default:
				log.Println("log queue is full, drop log")
			}
		}(reqLog, content)

		return nil
	}
}
