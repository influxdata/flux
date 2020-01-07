#ifndef _INFLUXDATA_FLUX_H
#define _INFLUXDATA_FLUX_H

#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

// flux_buffer_t is a reference to a byte-slice.
struct flux_buffer_t {
	// data is a pointer to the data contained within the buffer.
	void *data;

	// len is the length of the buffer.
	size_t len;
};

// flux_error_t represents a flux error.
struct flux_error_t;

// flux_error_str will return a string representation of the error.
// This will allocate memory for the returned string.
const char *flux_error_str(struct flux_error_t *);

// flux_free will free a resource.
void flux_free(void *);

// flux_ast_pkg_t is the AST representation of a flux query as a package.
struct flux_ast_pkg_t;

// flux_parse will take in a string and return the AST representation
// of the query.
struct flux_ast_pkg_t *flux_parse(const char *);

// flux_ast_marshal_json will marshal json and fill in the given buffer
// with the data. If successful, memory will be allocated for the data
// within the buffer and it is the caller's responsibility to free this
// data. If an error happens it will be returned. The error must be freed
// using flux_free if it is non-null.
struct flux_error_t *flux_ast_marshal_json(struct flux_ast_pkg_t *, struct flux_buffer_t *);

// flux_ast_marshal_fb will marshal the given AST as a flatbuffer into
// the given buffer. If successful, memory will be allocated for the data
// within the buffer and it is the caller's responsibility to free this
// data. If an error happens it will be returned. The error must be freed
// using flux_free if it is non-null.
struct flux_error_t *flux_ast_marshal_fb(struct flux_ast_pkg_t *, struct flux_buffer_t *);

// flux_semantic_pkg_t represents a semantic graph package node, including all of its files
// and their contents.
struct flux_semantic_pkg_t;

// flux_analyze analyzes the given AST and will populate the second pointer argument with
// a pointer to the resulting semantic graph. It is the caller's responsibility to free the
// resulting semantic graph with a call to flux_free().
// If analysis fails, the second pointer argument wil be pointed at 0, and an error will be returned.
// Any non-null error must be freed by calling flux_free.
// Regardless of whether an error is returned, this function will consume and free its
// flux_ast_pkg_t* argument, so it should not be reused after calling this function.
struct flux_error_t *flux_analyze(struct flux_ast_pkg_t *, struct flux_semantic_pkg_t **);

// flux_semantic_marshal_fb will marshal the given semantic graph as a flatbuffer into
// the given buffer. If successful, memory will be allocated for the data
// within the buffer and it is the caller's responsibility to free this
// data. If an error happens it will be returned. The error must be freed
// using flux_free if it is non-null.
struct flux_error_t *flux_semantic_marshal_fb(struct flux_semantic_pkg_t *, struct flux_buffer_t *);

#ifdef __cplusplus
}
#endif

#endif
