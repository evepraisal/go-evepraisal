package web

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/mash/go-accesslog"
)

type accessLogger struct {
}

func (l accessLogger) Log(record accesslog.LogRecord) {
	// Using the nginx combined access log format:
	// log_format combined '$remote_addr - $remote_user [$time_local] '
	//                    '"$request" $status $body_bytes_sent '
	//                    '"$http_referer" "$http_user_agent"';

	var buffer bytes.Buffer
	buffer.WriteString(emptyDash(record.Ip))
	buffer.WriteString(" - ")
	buffer.WriteString(emptyDash(record.RequestHeader.Get("logged-in-user")))
	buffer.WriteString(" ")
	buffer.WriteString("[" + record.Time.Format("02/Jan/2006:15:04:05 -0700") + "]")
	buffer.WriteString(" \"")
	buffer.WriteString(record.Method)
	buffer.WriteString(" ")
	buffer.WriteString(record.Uri)
	buffer.WriteString(" ")
	buffer.WriteString(record.Protocol)
	buffer.WriteString("\" ")
	buffer.WriteString(strconv.FormatInt(int64(record.Status), 10))
	buffer.WriteString(" ")
	buffer.WriteString(strconv.FormatInt(record.Size, 10))
	buffer.WriteString(" ")
	buffer.WriteString("\"" + emptyDash(record.RequestHeader.Get("Referer")) + "\"")
	buffer.WriteString(" ")
	buffer.WriteString("\"" + emptyDash(record.RequestHeader.Get("User-Agent")) + "\"")
	fmt.Println("access: " + buffer.String())
}

func emptyDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
