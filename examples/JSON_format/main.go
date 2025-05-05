package main

import (
	"encoding/json"
	"fmt"

	hqgoerrors "github.com/hueristiq/hq-go-errors"
)

func main() {
	err := hqgoerrors.New("root error example!", hqgoerrors.WithType("EXAMPLE_TYPE"), hqgoerrors.WithField("FIELD_KEY", "FIELD_VALUE"))
	err = hqgoerrors.Wrap(err, "wrapped error example!")

	formattedJSON := hqgoerrors.ToJSON(err, true)

	bytes, _ := json.Marshal(formattedJSON)

	fmt.Println(string(bytes))
}
