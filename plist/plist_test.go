package plist

import (
	"testing"
)

/*
Hmm, had to write this because of the following problem.
The key mapping uses large uint and terminal requires the formt 0x700000xx
This is the command line that work

$> hidutil property --set '{"UserKeyMapping":[{"HIDKeyboardModifierMappingSrc":0x7000000E7,"HIDKeyboardModifierMappingDst":0x7000000E0}]}'

If i use a map for that, the values are converted to decimal uint64 and that doesn't work
If I set as string then by converting the map to json it adds quotes around value "0x700000xx"

Instead I should do my own function to create the command to add to plist, hence the testing
*/

func TestMappingTemplate(t *testing.T) {
	tests := []struct {
		input  []Swappable
		output string
	}{
		{[]Swappable{}, `{"UserKeyMapping":[]}`},
		{[]Swappable{{HIDKeyboardModifierMappingSrc: 30064771296, HIDKeyboardModifierMappingDst: 30064771303}}, `{"UserKeyMapping":[ {"HIDKeyboardModifierMappingSrc":0x7000000e0,"HIDKeyboardModifierMappingDst":0x7000000e7}]}`},
	}

	for _, test := range tests {
		result := createUserKeyMapping(test.input)
		if result != test.output {
			t.Errorf("Expected \n%q \n\t-- but got --\n%q\n", test.output, result)
		}
	}
}

func TestSwappableIsComplete(t *testing.T) {
	tests := []struct {
		input  Swappable
		output bool
	}{
		{input: Swappable{HIDKeyboardModifierMappingSrc: 30064771303}, output: false},
		{input: Swappable{HIDKeyboardModifierMappingSrc: 30064771303, HIDKeyboardModifierMappingDst: 30064771296}, output: true},
	}

	for _, test := range tests {
		result := test.input.isComplete()
		if result != test.output {
			t.Errorf("For input %+v\n expected %t but got %t\n", test.input, test.output, result)
		}
	}
}
