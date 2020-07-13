package safehttp

// Machinery TODO
type Machinery struct {
	h HandleFunc
	d Dispatcher
}

// NewMachinery TODO
func NewMachinery(h HandleFunc, d Dispatcher) *Machinery {
	return &Machinery{h: h, d: d}
}

// HandleRequest TODO
func (m *Machinery) HandleRequest(r string) {
	rw := &ResponseWriter{d: m.d}
	m.h(rw, nil)
}
