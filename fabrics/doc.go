// Package fabrics wraps libmxl-fabrics, the cross-node transport
// layer of libmxl. It exposes Instance, Target, and Initiator
// handles, the Provider enum (auto/tcp/verbs/efa/shm), and helpers
// to register a FlowReader's or FlowWriter's shared memory as a
// libmxl-fabrics memory region.
//
// All handles use explicit Close and a runtime finalizer that calls
// Close on garbage collection, matching the convention of the parent
// mxl package. Methods are safe for concurrent use unless documented
// otherwise.
package fabrics
