// Package fabrics wraps libmxl-fabrics, the cross-node transport
// layer of libmxl. It exposes Instance, Target, and Initiator
// handles and the Provider enum (any/tcp/verbs/efa/shm). A Target
// receives grains or samples into a FlowWriter; an Initiator sends a
// FlowReader's grains or samples to one or more targets.
//
// All handles use explicit Close and a runtime finalizer that calls
// Close on garbage collection, matching the convention of the parent
// mxl package. Methods are safe for concurrent use unless documented
// otherwise.
package fabrics
