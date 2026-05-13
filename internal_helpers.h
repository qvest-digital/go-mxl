/* SPDX-FileCopyrightText: 2025 Contributors to the Media eXchange Layer project.
 * SPDX-License-Identifier: Apache-2.0
 *
 * Tiny cgo helpers shared across the Go binding. cgo gives each *.go file its
 * own C translation unit, so static inline helpers defined in one file's
 * preamble are not visible in another. Centralizing them here and #include'ing
 * from every preamble keeps the helpers in lockstep across files.
 *
 * Two kinds of helpers live here:
 *   - typed pointers into mxlFlowConfigInfo's anonymous union, which cgo
 *     cannot address directly;
 *   - fragment / stride / count accessors for the wrapped multi-buffer slice
 *     types used by the continuous sample APIs, which cgo also cannot index
 *     ergonomically because they are fixed-size arrays inside nested structs.
 */
#pragma once

#include <stddef.h>
#include <mxl/flow.h>
#include <mxl/flowinfo.h>

static inline mxlDiscreteFlowConfigInfo const* mxl_go_config_discrete(mxlFlowConfigInfo const* c) {
    return &c->discrete;
}

static inline mxlContinuousFlowConfigInfo const* mxl_go_config_continuous(mxlFlowConfigInfo const* c) {
    return &c->continuous;
}

static inline void const* mxl_go_wrapped_fragment_ptr_ro(mxlWrappedMultiBufferSlice const* s, size_t i) {
    return s->base.fragments[i].pointer;
}

static inline size_t mxl_go_wrapped_fragment_size_ro(mxlWrappedMultiBufferSlice const* s, size_t i) {
    return s->base.fragments[i].size;
}

static inline size_t mxl_go_wrapped_stride_ro(mxlWrappedMultiBufferSlice const* s) {
    return s->stride;
}

static inline size_t mxl_go_wrapped_count_ro(mxlWrappedMultiBufferSlice const* s) {
    return s->count;
}

static inline void* mxl_go_wrapped_fragment_ptr_rw(mxlMutableWrappedMultiBufferSlice const* s, size_t i) {
    return s->base.fragments[i].pointer;
}

static inline size_t mxl_go_wrapped_fragment_size_rw(mxlMutableWrappedMultiBufferSlice const* s, size_t i) {
    return s->base.fragments[i].size;
}

static inline size_t mxl_go_wrapped_stride_rw(mxlMutableWrappedMultiBufferSlice const* s) {
    return s->stride;
}

static inline size_t mxl_go_wrapped_count_rw(mxlMutableWrappedMultiBufferSlice const* s) {
    return s->count;
}
