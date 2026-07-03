// Package identity defines node recognition for the coordination protocol: the
// node's self-chosen identifier and the single canonical way a caller's
// workload identity is bound to it. Sovereignty of the node begins here — the
// node names itself, and no coordinator speaks in its place.
package identity

// NodeID is the stable identifier a node uses for itself across coordinators.
type NodeID string

// SPIFFEPathForNode returns the SPIFFE ID path a node's workload identity MUST
// present. This is the one canonical binding format: an SVID is issued only
// under this path, and a coordinator authorizes a request as coming from a node
// by comparing the caller's SPIFFE path to this value. It is byte-identical to
// the format SoHoLINK already enforces in production (deterministic
// construction, never a per-node database lookup).
func SPIFFEPathForNode(n NodeID) string {
	return "/node/" + string(n)
}

// BindsTo reports whether a caller presenting spiffePath is authorized to act
// as node n.
//
// A conformant coordinator MUST reject with 403 when this returns false, and
// MUST reject with 401 when no SPIFFE identity is present at all. There is no
// "any valid SVID" authorization path: accepting any node's SVID for any node's
// request is an impersonation vector, and a check that is present but inert
// (gated on a field that is never populated) is worse than an absent one.
func BindsTo(spiffePath string, n NodeID) bool {
	return spiffePath == SPIFFEPathForNode(n)
}
