package null

import "encoding/json"

type String struct {
	Value    string
	Explicit bool
	IsNull   bool
}

type Int struct {
	Value    int
	Explicit bool
	IsNull   bool
}

type Bool struct {
	Value    bool
	Explicit bool
	IsNull   bool
}

type Array[T any] struct {
	Value    []T
	Explicit bool
	IsNull   bool
}

func (ns *String) UnmarshalJSON(data []byte) error {
	ns.Explicit = true

	if string(data) == "null" {
		ns.IsNull = true
		ns.Value = ""
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	ns.IsNull = false
	ns.Value = s
	return nil
}

func (ns *Array[T]) UnmarshalJSON(data []byte) error {
	ns.Explicit = true

	if string(data) == "null" {
		ns.IsNull = true
		ns.Value = []T{}
		return nil
	}

	var s []T
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	ns.IsNull = false
	ns.Value = s
	return nil
}

func (ni *Int) UnmarshalJSON(data []byte) error {
	ni.Explicit = true

	if string(data) == "null" {
		ni.IsNull = true
		ni.Value = 0
		return nil
	}

	var i int
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}

	ni.IsNull = false
	ni.Value = i
	return nil
}

func (nb *Bool) UnmarshalJSON(data []byte) error {
	nb.Explicit = true

	if string(data) == "null" {
		nb.IsNull = true
		nb.Value = false
		return nil
	}

	var b bool
	if err := json.Unmarshal(data, &b); err != nil {
		return err
	}

	nb.IsNull = false
	nb.Value = b
	return nil
}
