package cli

import (
	"strconv"
)

type Value interface {
	String() string
	Set(string) error
	Get() any
}

type boolValue bool

func (v *boolValue) String() string {
	return strconv.FormatBool(bool(*v))
}

func (v *boolValue) Set(s string) error {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}

	*v = boolValue(b)
	return nil
}

func (v *boolValue) Get() any {
	return bool(*v)
}

func newBoolValue(b bool) *boolValue {
	return (*boolValue)(&b)
}

type Flag interface {
	Name() (string, string)
	Description() string
	Value() Value
}

type BoolFlag struct {
	long, short, description string
	value                    Value
}

func (f *BoolFlag) Name() (string, string) {
	return f.long, f.short
}

func (f *BoolFlag) Description() string {
	return f.description
}

func (f *BoolFlag) Value() Value {
	return f.value
}

func NewBoolFlagP(long, short string, defVal bool, description string) Flag {
	return &BoolFlag{long, short, description, newBoolValue(defVal)}
}

type FlagSet map[string]Flag

func (s FlagSet) GetBool(long string) bool {
	f, ok := s[long]
	if !ok {
		return false
	}

	b, ok := f.Value().Get().(bool)
	if !ok {
		return false
	}

	return b
}
