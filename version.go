// Package sohocloud is the root of the sohocloud-protocol module: the shared,
// language-of-record definition of the substrate COORDINATION protocol —
// node recognition, capability listing, and job employment — spoken uniformly
// by every coordinator and every frontend on the network.
//
// Scope boundary: this module governs COORDINATION ONLY. It is not the JFA
// member economy. Escrow, member-issued credit, the reputation covenant, the
// dialog-sealed record, and dispute adjudication belong to each frontend and
// are deliberately absent here. Nothing in this module imports any other NTARI
// module; it is a dependency leaf, so that no coordinator (SoHoLINK included)
// can become a hub every consumer must route through (open problem #7).
package sohocloud

// Version is the protocol version. v0 is UNSTABLE: message shapes and the
// canonical byte format may change without compatibility guarantees until v1.
const Version = "v0"
