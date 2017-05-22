package web

import (
	"log"

	"github.com/mash/go-accesslog"
)

type accessLogger struct {
}

func (l accessLogger) Log(record accesslog.LogRecord) {
	log.Printf("%s %s%s %d (%s) - %db, %s",
		record.Method, record.Host, record.Uri, record.Status, record.Ip, record.Size, record.ElapsedTime)
}
