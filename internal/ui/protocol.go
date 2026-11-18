package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"

	"github.com/yuwen002/meteor-ch-tui/internal/ch"
)

func nextProtocol(cur string) string {
	switch ch.NormalizeProtocol(cur) {
	case ch.ProtocolNative:
		return ch.ProtocolHTTP
	case ch.ProtocolHTTP:
		return ch.ProtocolHTTPS
	default:
		return ch.ProtocolNative
	}
}

func defaultPortFor(protocol string) string {
	switch ch.NormalizeProtocol(protocol) {
	case ch.ProtocolHTTP:
		return "8123"
	case ch.ProtocolHTTPS:
		return "8443"
	default:
		return "9000"
	}
}

func applyProtocolPort(ti *textinput.Model, protocol string) {
	port := strings.TrimSpace(ti.Value())
	switch port {
	case "", "9000", "8123", "8443":
		ti.SetValue(defaultPortFor(protocol))
	}
}

func protocolPortNote(protocol, portStr string) string {
	port := strings.TrimSpace(portStr)
	def := defaultPortFor(protocol)
	if port == "" || port == def {
		return ""
	}
	return fmt.Sprintf("  ! 端口 %s（默认 %s）", port, def)
}

func (m *Model) toggleMgrAddProtocol() {
	next := nextProtocol(m.mgrAddProtocol.Value())
	m.mgrAddProtocol.SetValue(next)
	applyProtocolPort(&m.mgrAddPort, next)
}

func (m *Model) toggleCfgProtocol() {
	ti := &m.cfgFields[cfgFieldProtocol]
	next := nextProtocol(ti.Value())
	ti.SetValue(next)
	applyProtocolPort(&m.cfgFields[cfgFieldPort], next)
}
