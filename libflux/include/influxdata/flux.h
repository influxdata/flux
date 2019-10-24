#ifndef _INFLUXDATA_FLUX_H
#define _INFLUXDATA_FLUX_H

#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

// flux_ast_t is the AST representation of a flux query.
struct flux_ast_t;

// flux_buffer_t is a reference to a byte-slice.
struct flux_buffer_t {
	// data is a pointer to the data contained within the buffer.
	void *data;

	// len is the length of the buffer.
	size_t len;
};

// flux_error_t represents a flux error.
struct flux_error_t;

// flux_parse will take in a string and return the AST representation
// of the query.
struct flux_ast_t *flux_parse(const char *);

// flux_ast_marshal_json will marshal json and fill in the given buffer
// with the data. If successful, memory will be allocated for the data
// within the buffer and it is the caller's responsibility to free this
// data. If an error happens it will be returned. The error must be freed
// using flux_free if it is non-null.
struct flux_error_t *flux_ast_marshal_json(struct flux_ast_t *, struct flux_buffer_t *);

// flux_buffer_free will free the memory that was allocated for a buffer.
// This should only be called if the caller is the one who owns the data.
void flux_buffer_free(struct flux_buffer_t *);

// flux_error_str will return a string representation of the error.
// This will allocate memory for the returned string.
const char *flux_error_str(struct flux_error_t *);

// flux_free will free a resource.
void flux_free(void *);

#ifdef __cplusplus
}
#endif

#endif
