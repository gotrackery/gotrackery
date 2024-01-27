package internal

const device = "device"

// Session is a session context storage. It is used to store context that must be shared during whole session.
// Device id for instance.
type Session struct {
	ctx map[string]interface{}
}

// NewSession creates a new session context storage.
func NewSession() *Session {
	return &Session{
		ctx: make(map[string]interface{}),
	}
}

// Set sets a value for a given key.
func (s *Session) Set(key string, value interface{}) {
	if s.ctx == nil {
		s.ctx = make(map[string]interface{})
	}
	s.ctx[key] = value
}

// Get returns the value associated with the given key.
func (s *Session) Get(key string) interface{} {
	if s.ctx == nil {
		return nil
	}
	return s.ctx[key]
}

// SetDevice sets the device id to session context.
func (s *Session) SetDevice(dev string) {
	if s.ctx == nil {
		s.ctx = make(map[string]interface{})
	}
	s.ctx[device] = dev
}

// GetDevice returns the device id from session context.
func (s *Session) Device() string {
	if s.ctx == nil {
		return ""
	}
	if d, ok := s.ctx[device]; ok {
		return d.(string)
	}
	return ""
}
