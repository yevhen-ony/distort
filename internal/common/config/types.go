package config

import "github.com/docker/go-units"

type Size int64

func (s *Size) UnmarshalText(text []byte) error {
	v, err := units.FromHumanSize(string(text))
	if err != nil {
		return err
	}
	*s = Size(v)
	return nil
}
