package common

import (
	"testing"
)


func TestStoreErr(t *testing.T) {
	E := NewStoreErr("test", KeyNotFound, "0")
	if E.Error() != "test, 0, Not Found" {
		t.Errorf("Expected 'test, 0, Not Found' but found '%v'", E.Error())
	}
	E = NewStoreErr("test", TooLate, "0")
	if E.Error() != "test, 0, Too Late" {
		t.Errorf("Expected 'test, 0, Too Late' but found '%v'", E.Error())
	}
	E = NewStoreErr("test", PassedIndex, "0")
	if E.Error() != "test, 0, Passed Index" {
		t.Errorf("Expected 'test, 0, Passed Index' but found '%v'", E.Error())
	}
	E = NewStoreErr("test", SkippedIndex, "0")
	if E.Error() != "test, 0, Skipped Index" {
		t.Errorf("Expected 'test, 0, Skipped Index' but found '%v'", E.Error())
	}
	E = NewStoreErr("test", NoRoot, "0")
	if E.Error() != "test, 0, No Root" {
		t.Errorf("Expected 'test, 0, No Root' but found '%v'", E.Error())
	}
	E = NewStoreErr("test", UnknownParticipant, "0")
	if E.Error() != "test, 0, Unknown Participant" {
		t.Errorf("Expected 'test, 0, Unknown Participant' but found '%v'", E.Error())
	}
	E = NewStoreErr("test", Empty, "0")
	if E.Error() != "test, 0, Empty" {
		t.Errorf("Expected 'test, 0, Empty' but found '%v'", E.Error())
	}
	E = NewStoreErr("test", UnknownParticipant, "0")
}
