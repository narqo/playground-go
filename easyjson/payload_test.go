package easyjson

import (
	"encoding/json"
	"fmt"
)

func ExamplePayload_UnmarshalJSON() {
	rawJSON := `{"apiKey": "123", "events": {}}`

	var payload Payload
	err := json.Unmarshal([]byte(rawJSON), &payload)
	fmt.Println(err)

	// Output:
}
