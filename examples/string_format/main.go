package main

import (
	"fmt"

	hqgoerrors "github.com/hueristiq/hq-go-errors"
)

func main() {
	err := hqgoerrors.New("root error example!", hqgoerrors.WithType("EXAMPLE_TYPE"), hqgoerrors.WithField("FIELD_KEY", "FIELD_VALUE"))
	err = hqgoerrors.Wrap(err, "wrapped error example!")

	formattedStr := hqgoerrors.ToString(err, true)

	fmt.Println(formattedStr)
}
