package extract

import "testing"

func TestSetGetDefuddleJS(t *testing.T) {
	original := GetDefuddleJS()
	defer SetDefuddleJS(original) // restore after test

	js := "function defuddle() { return 'test'; }"
	SetDefuddleJS(js)

	got := GetDefuddleJS()
	if got != js {
		t.Errorf("GetDefuddleJS() = %q, want %q", got, js)
	}
}
