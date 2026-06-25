package handler

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// jsonString accepts OTP code as JSON string or number (Postman often sends number).
type jsonString string

func (s *jsonString) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = jsonString(str)
		return nil
	}

	var num json.Number
	if err := json.Unmarshal(data, &num); err == nil {
		*s = jsonString(num.String())
		return nil
	}

	var n int64
	if err := json.Unmarshal(data, &n); err == nil {
		*s = jsonString(strconv.FormatInt(n, 10))
		return nil
	}

	return fmt.Errorf("value must be string or number")
}

func (s jsonString) String() string {
	return string(s)
}
