package main

import (
	hqgoerrors "github.com/hueristiq/hq-go-errors"
	hqgologger "github.com/hueristiq/hq-go-logger"
)

func main() {
	err := hqgoerrors.New("root error example!", hqgoerrors.WithType("ERROR_TYPE"), hqgoerrors.WithField("FIELD_KEY_1", "FIELD_VALUE_1"), hqgoerrors.WithField("FIELD_KEY_2", "FIELD_VALUE_2"))

	err = hqgoerrors.Wrap(err, "wrap error example 1!")
	err = hqgoerrors.Wrap(err, "wrap error example 2!", hqgoerrors.WithType("ERROR_TYPE_2"), hqgoerrors.WithField("FIELD_KEY_1", "FIELD_VALUE_1"), hqgoerrors.WithField("FIELD_KEY_2", "FIELD_VALUE_2"))

	formattedStr := hqgoerrors.ToString(err, true)

	hqgologger.Error().Label("").Msg(formattedStr)
}
