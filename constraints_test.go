package semver

import (
	"reflect"
	"testing"
)

func TestParseConstraint(t *testing.T) {
	tests := []struct {
		in  string
		f   cfunc
		v   string
		err bool
	}{
		{">= 1.2", constraintGreaterThanEqual, "1.2.0", false},
		{"1.0", constraintEqual, "1.0.0", false},
		{"foo", nil, "", true},
		{"<= 1.2", constraintLessThanEqual, "1.2.0", false},
		{"=< 1.2", constraintLessThanEqual, "1.2.0", false},
		{"=> 1.2", constraintGreaterThanEqual, "1.2.0", false},
		{"v1.2", constraintEqual, "1.2.0", false},
		{"=1.5", constraintEqual, "1.5.0", false},
		{"> 1.3", constraintGreaterThan, "1.3.0", false},
		{"< 1.4.1", constraintLessThan, "1.4.1", false},
	}

	for _, tc := range tests {
		c, err := parseConstraint(tc.in)
		if tc.err && err == nil {
			t.Errorf("Expected error for %s didn't occur", tc.in)
		} else if !tc.err && err != nil {
			t.Errorf("Unexpected error for %s", tc.in)
		}

		// If an error was expected continue the loop and don't try the other
		// tests as they will cause errors.
		if tc.err {
			continue
		}

		if tc.v != c.con.String() {
			t.Errorf("Incorrect version found on %s", tc.in)
		}

		f1 := reflect.ValueOf(tc.f)
		f2 := reflect.ValueOf(c.function)
		if f1 != f2 {
			t.Errorf("Wrong constraint found for %s", tc.in)
		}
	}
}

func TestConstraintCheck(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		check      bool
	}{
		{"= 2.0", "1.2.3", false},
		{"= 2.0", "2.0.0", true},
		{"4.1", "4.1.0", true},
		{"!=4.1", "4.1.0", false},
		{"!=4.1", "5.1.0", true},
		{">1.1", "4.1.0", true},
		{">1.1", "1.1.0", false},
		{"<1.1", "0.1.0", true},
		{"<1.1", "1.1.0", false},
		{"<1.1", "1.1.1", false},
		{">=1.1", "4.1.0", true},
		{">=1.1", "1.1.0", true},
		{">=1.1", "0.0.9", false},
		{"<=1.1", "0.1.0", true},
		{"<=1.1", "1.1.0", true},
		{"<=1.1", "1.1.1", false},
	}

	for _, tc := range tests {
		c, err := parseConstraint(tc.constraint)
		if err != nil {
			t.Errorf("err: %s", err)
			continue
		}

		v, err := NewVersion(tc.version)
		if err != nil {
			t.Errorf("err: %s", err)
			continue
		}

		a := c.check(v)
		if a != tc.check {
			t.Errorf("Constraint '%s' failing", tc.constraint)
		}
	}
}

func TestNewConstraint(t *testing.T) {
	tests := []struct {
		input string
		ors   int
		count int
		err   bool
	}{
		{">= 1.1", 1, 1, false},
		{"2.0", 1, 1, false},
		{">= bar", 0, 0, true},
		{">= 1.2.3, < 2.0", 1, 2, false},
		{">= 1.2.3, < 2.0 || => 3.0, < 4", 2, 2, false},

		// The 3-4 should be broken into 2 by the range rewriting
		{"3-4 || => 3.0, < 4", 2, 2, false},
	}

	for _, tc := range tests {
		v, err := NewConstraint(tc.input)
		if tc.err && err == nil {
			t.Errorf("expected but did not get error for: %s", tc.input)
			continue
		} else if !tc.err && err != nil {
			t.Errorf("unexpectederror for input %s: %s", tc.input, err)
			continue
		}
		if tc.err {
			continue
		}

		l := len(v.constraints)
		if tc.ors != l {
			t.Errorf("Expected %s to have %d ORs but got %d",
				tc.input, tc.ors, l)
		}

		l = len(v.constraints[0])
		if tc.count != l {
			t.Errorf("Expected %s to have %d constraints but got %d",
				tc.input, tc.count, l)
		}
	}
}

func TestConstraintsCheck(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		check      bool
	}{
		{"= 2.0", "1.2.3", false},
		{"= 2.0", "2.0.0", true},
		{"4.1", "4.1.0", true},
		{"!=4.1", "4.1.0", false},
		{"!=4.1", "5.1.0", true},
		{">1.1", "4.1.0", true},
		{">1.1", "1.1.0", false},
		{"<1.1", "0.1.0", true},
		{"<1.1", "1.1.0", false},
		{"<1.1", "1.1.1", false},
		{">=1.1", "4.1.0", true},
		{">=1.1", "1.1.0", true},
		{">=1.1", "0.0.9", false},
		{"<=1.1", "0.1.0", true},
		{"<=1.1", "1.1.0", true},
		{"<=1.1", "1.1.1", false},
		{">1.1, <2", "1.1.1", true},
		{">1.1, <3", "4.3.2", false},
		{">=1.1, <2, !=1.2.3", "1.2.3", false},
		{">=1.1, <2, !=1.2.3 || > 3", "3.1.2", true},
		{">=1.1, <2, !=1.2.3 || >= 3", "3.0.0", true},
		{">=1.1, <2, !=1.2.3 || > 3", "3.0.0", false},
		{">=1.1, <2, !=1.2.3 || > 3", "1.2.3", false},
		{"1.1 - 2", "1.1.1", true},
		{"1.1-3", "4.3.2", false},
		{"^1.1", "1.1.1", true},
		{"^1.1", "4.3.2", false},
		{"^1.x", "1.1.1", true},
		{"^1.x", "2.1.1", false},
	}

	for _, tc := range tests {
		c, err := NewConstraint(tc.constraint)
		if err != nil {
			t.Errorf("err: %s", err)
			continue
		}

		v, err := NewVersion(tc.version)
		if err != nil {
			t.Errorf("err: %s", err)
			continue
		}

		a := c.Check(v)
		if a != tc.check {
			t.Errorf("Constraint '%s' failing", tc.constraint)
		}
	}
}

func TestRewriteRange(t *testing.T) {
	tests := []struct {
		c  string
		nc string
	}{
		{"2-3", ">= 2, <= 3"},
		{"2-3, 2-3", ">= 2, <= 3,>= 2, <= 3"},
		{"2-3, 4.0.0-5.1", ">= 2, <= 3,>= 4.0.0, <= 5.1"},
	}

	for _, tc := range tests {
		o := rewriteRange(tc.c)

		if o != tc.nc {
			t.Errorf("Range %s rewritten incorrectly as '%s'", tc.c, o)
		}
	}
}

func TestIsX(t *testing.T) {
	tests := []struct {
		t string
		c bool
	}{
		{"A", false},
		{"%", false},
		{"X", true},
		{"x", true},
		{"*", true},
	}

	for _, tc := range tests {
		a := isX(tc.t)
		if a != tc.c {
			t.Errorf("Function isX error on %s", tc.t)
		}
	}
}

func TestRewriteCarets(t *testing.T) {
	tests := []struct {
		c  string
		nc string
	}{
		{"1.1, 4.0.0 - 5.1", "1.1, 4.0.0 - 5.1"},
		{"^*", ">=0.0.0"},
		{"^x", ">=0.0.0"},
		{"^2", ">= 2, < 3"},
		{"^2, ^2", ">= 2, < 3, >= 2, < 3"},
		{"^2.1", ">= 2.1, < 3"},
		{"^2.1.3", ">= 2.1.3, < 3"},
		{"^1.1, 4.0.0 - 5.1", ">= 1.1, < 2, 4.0.0 - 5.1"},
		{"^1.x", ">= 1.0, < 2"},
		{"^1.2.x", ">= 1.2.0, < 2"},
		{"^1.2.x-beta.1+foo", ">= 1.2.0-beta.1, < 2"},
	}

	for _, tc := range tests {
		o := rewriteCarets(tc.c)

		if o != tc.nc {
			t.Errorf("Carets %s rewritten incorrectly as '%s'", tc.c, o)
		}
	}
}