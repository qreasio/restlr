package model

// Option represents a row from 'options'.
type Option struct {
	OptionID    uint64 `json:"option_id"`    // option_id
	OptionName  string `json:"option_name"`  // option_name
	OptionValue string `json:"option_value"` // option_value
	AutoLoad    string `json:"autoload"`     // autoload
}
