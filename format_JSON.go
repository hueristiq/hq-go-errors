package errors

// JSONFormat holds formatting rules for serializing errors into JSON-like maps.
// The StackElemSep separates method, file, and line in each frame string.
//
// Fields:
//   - Options (FormatOptions): Format options (e.g. omitting stack trace or inverting the output order).
//   - StackElemSep (string): Separator between elements of each stack frame.
type JSONFormat struct {
	Options      FormatOptions
	StackElemSep string
}

// NewDefaultJSONFormat returns a JSONFormat with default separators.
// StackElemSep defaults to ":".
func NewDefaultJSONFormat(options FormatOptions) (format JSONFormat) {
	format = JSONFormat{
		Options:      options,
		StackElemSep: ":",
	}

	return
}

// ToJSON returns a JSON formatted map for a given error.
//
// An error without trace will be formatted as follows:
//
//	{
//	  "root": {
//	      "message": "Root error msg"
//	  },
//	  "wrap": [
//	    {
//	      "message": "Wrap error msg"
//	    }
//	  ]
//	}
//
// An error with trace will be formatted as follows:
//
//	{
//	  "root": {
//	    "message": "Root error msg",
//	    "stack": [
//	      "<Method2>:<File2>:<Line2>",
//	      "<Method1>:<File1>:<Line1>"
//	    ]
//	  },
//	  "wrap": [
//	    {
//	      "message": "Wrap error msg",
//	      "stack": "<Method2>:<File2>:<Line2>"
//	    }
//	  ]
//	}
func ToJSON(err error, withTrace bool) map[string]interface{} {
	return ToCustomJSON(err, NewDefaultJSONFormat(FormatOptions{
		WithTrace:    withTrace,
		WithExternal: true,
	}))
}

// ToCustomJSON returns a JSON formatted map for a given error.
//
// To declare custom format, the Format object has to be passed as an argument.
// An error without trace will be formatted as follows:
//
//	{
//	  "root": {
//	    "message": "Root error msg",
//	  },
//	  "wrap": [
//	    {
//	      "message": "Wrap error msg'",
//	    }
//	  ]
//	}
//
// An error with trace will be formatted as follows:
//
//	{
//	  "root": {
//	    "message": "Root error msg",
//	    "stack": [
//	      "<Method2>[Format.StackElemSep]<File2>[Format.StackElemSep]<Line2>",
//	      "<Method1>[Format.StackElemSep]<File1>[Format.StackElemSep]<Line1>"
//	    ]
//	  }
//	  "wrap": [
//	    {
//	      "message": "Wrap error msg",
//	      "stack": "<Method2>[Format.StackElemSep]<File2>[Format.StackElemSep]<Line2>"
//	    }
//	  ]
//	}
func ToCustomJSON(err error, format JSONFormat) map[string]interface{} {
	upErr := Unpack(err)

	jsonMap := make(map[string]interface{})

	if format.Options.WithExternal && upErr.ErrExternal != nil {
		jsonMap["external"] = formatExternalStr(upErr.ErrExternal, format.Options.WithTrace)
	}

	if upErr.ErrRoot.Msg != "" || len(upErr.ErrRoot.Stack) > 0 {
		jsonMap["root"] = upErr.ErrRoot.formatJSON(format)
	}

	if len(upErr.ErrChain) > 0 {
		var wrapArr []map[string]interface{}

		for _, eLink := range upErr.ErrChain {
			wrapMap := eLink.formatJSON(format)

			if format.Options.InvertOutput {
				wrapArr = append(wrapArr, wrapMap)
			} else {
				wrapArr = append([]map[string]interface{}{wrapMap}, wrapArr...)
			}
		}

		jsonMap["wrap"] = wrapArr
	}

	return jsonMap
}
