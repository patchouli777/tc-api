package sl

import "log/slog"

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

// TODO: slog.String()
func Str(key, val string) slog.Attr {
	return slog.Attr{
		Key:   key,
		Value: slog.StringValue(val),
	}
}

// TODO: slog.Bool
func Bool(key string, val bool) slog.Attr {
	return slog.Attr{
		Key:   key,
		Value: slog.BoolValue(val),
	}
}
