package extract

// DefuddleJS is set at initialization time from the embedded JS.
// The actual //go:embed happens in cmd/rlss/ where the relative path works.
var defuddleJS string

// SetDefuddleJS sets the Defuddle JavaScript code.
// Must be called before creating a Pool.
func SetDefuddleJS(js string) {
	defuddleJS = js
}

// GetDefuddleJS returns the Defuddle JavaScript code.
func GetDefuddleJS() string {
	return defuddleJS
}
