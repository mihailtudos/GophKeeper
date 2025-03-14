package pkg

import "log/slog"

func ErrAttr(err error) slog.Attr {
	return slog.String("error", err.Error())
}
