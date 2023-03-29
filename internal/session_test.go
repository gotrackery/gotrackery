package internal

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestSession(t *testing.T) {
	s := &Session{}
	assert.Equal(t, s.Get("test"), nil, "Get should return nil")
	s.Set("test", "test")
	assert.Equal(t, s.Get("test"), "test", "Get should return test")

	s = &Session{}
	assert.Equal(t, s.GetDevice(), "", "GetDevice should return empty string, if not set")
	s.SetDevice("device")
	assert.Equal(t, s.GetDevice(), "device", "GetDevice should return device, if set")
	s = NewSession()
	assert.Equal(t, s.GetDevice(), "", "GetDevice should return empty string")
	s.SetDevice("device")
	assert.Equal(t, s.GetDevice(), "device", "GetDevice should return device")

	f := func(session *Session) {
		session.Set("internal", "test")
		session.Set("test", "tested")
	}
	assert.Equal(t, s.Get("internal"), nil, "Get should return nil")
	f(s)
	assert.Equal(t, s.Get("internal"), "test", "Get should return internal")
	assert.Equal(t, s.Get("test"), "tested", "Get should return test")
}
