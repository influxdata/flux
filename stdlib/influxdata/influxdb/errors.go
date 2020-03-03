package influxdb

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

func handleError(target interface{}) error {
	if errStr, ok := target.(string); ok {
		return errors.New(codes.Unknown, errStr)
	}
	if internalErrMap, ok := target.(map[string]interface{}); ok {
		internalErr := new(flux.Error)
		if code, ok := internalErrMap["code"].(string); ok {
			internalErr.Code = handleErrorCode(code)
		}
		if msg, ok := internalErrMap["message"].(string); ok {
			internalErr.Msg = msg
		}
		if err, ok := internalErrMap["error"]; ok {
			internalErr.Err = handleError(err)
		}
		return internalErr
	}
	return nil
}

func handleErrorCode(code string) codes.Code {
	switch code {
	case "internal error":
		return codes.Internal
	case "not found":
		return codes.NotFound
	case "invalid":
		return codes.Invalid
	case "unavailable":
		return codes.Unavailable
	case "forbidden":
		return codes.PermissionDenied
	case "unauthorized":
		return codes.Unauthenticated
	default:
		return codes.Unknown
	}
}
