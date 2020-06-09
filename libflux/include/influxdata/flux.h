#ifndef _INFLUXDATA_FLUX_H
#define _INFLUXDATA_FLUX_H

#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

// flux_buffer_t is a reference to a byte-slice.
struct flux_buffer_t {
	// data is a pointer to the data contained within the buffer.
	char *data;

	// len is the length of the buffer.
	size_t len;
};

// flux_error_t represents a flux error.
struct flux_error_t;

// flux_free_error will release memory associated with an error.
void flux_free_error(struct flux_error_t *);

// flux_error_str will return a string representation of the error.
// This will allocate memory for the returned string, which must be
// freed wtih flux_free_bytes.
const char *flux_error_str(struct flux_error_t *);

// flux_free_bytes will release the memory pointed to by the pointer argument.
void flux_free_bytes(const char *);

// flux_ast_pkg_t is the AST representation of a flux query as a package.
struct flux_ast_pkg_t;

// flux_parse will take in a file name string and a source string then
// return the AST representation of the query.
struct flux_ast_pkg_t *flux_parse(const char *file_name, const char *flux_source);

// flux_ast_get_error will return the first error in the AST, if any.
struct flux_error_t *flux_ast_get_error(struct flux_ast_pkg_t *);

// flux_free_ast_pkg will release the memory associated with the given pointer.
void flux_free_ast_pkg(struct flux_ast_pkg_t *);

// flux_merge_ast_pkgs merges the files of a given input AST package into the file vector of a
// given output AST package. This function borrows the packages, but it does not own them. The
// caller of this function still needs to free the package memory on the Go side.
struct flux_error_t *flux_merge_ast_pkgs(struct flux_ast_pkg_t *, struct flux_ast_pkg_t *);

// flux_parse_json will take in a JSON string for an AST package
// and populate its second pointer argument with a pointer to an
// AST package.
// Note that the caller should free the pointer to the AST, not the pointer to the pointer
// to the AST.  It is the former that references memory allocated by Rust.
// If an error happens it will be returned. The error must be freed
// using flux_free_error if it is non-null.
struct flux_error_t *flux_parse_json(const char *, struct flux_ast_pkg_t **);

// flux_free_ast_pkg will release the memory associated with the given pointer.
void flux_free_ast_pkg(struct flux_ast_pkg_t *);

// flux_merge_ast_pkgs merges the files of a given input AST package into the file vector of a
// given output AST package. This function borrows the packages, but it does not own them. The
// caller of this function still needs to free the package memory on the Go side.
struct flux_error_t *flux_merge_ast_pkgs(struct flux_ast_pkg_t *, struct flux_ast_pkg_t *);

// flux_ast_marshal_json will marshal json and fill in the given buffer
// with the data. If successful, memory will be allocated for the data
// within the buffer and it is the caller's responsibility to free this
// data. If an error happens it will be returned. The error must be freed
// using flux_free_error if it is non-null.
struct flux_error_t *flux_ast_marshal_json(struct flux_ast_pkg_t *, struct flux_buffer_t *);

// flux_ast_marshal_fb will marshal the given AST as a flatbuffer into
// the given buffer. If successful, memory will be allocated for the data
// within the buffer and it is the caller's responsibility to free this
// data. If an error happens it will be returned. The error must be freed
// using flux_free_error if it is non-null.
struct flux_error_t *flux_ast_marshal_fb(struct flux_ast_pkg_t *, struct flux_buffer_t *);

// flux_get_env_stdlib instantiates a flatbuffers TypeEnvironment and creates a pointer
// to it to use when performing lookups on the stdlib
void flux_get_env_stdlib(struct flux_buffer_t *);

// flux_semantic_pkg_t represents a semantic graph package node, including all of its files
// and their contents.
struct flux_semantic_pkg_t;

// flux_semantic_analyzer_t represents a semantic analyzer that can be used to iteratively
// analyze snippets of code.
struct flux_semantic_analyzer_t;

// flux_new_semantic_analyzer creates a new semantic analyzer for the given package path.
// The returned analyzer must be freed with flux_free_semantic_analyzer().
struct flux_semantic_analyzer_t *flux_new_semantic_analyzer(const char *pkgpath);

// flux_free_semantic_analyzer frees a previously allocated semantic analyzer.
void flux_free_semantic_analyzer(struct flux_semantic_analyzer_t *);

// flux_analyze_with will analyze the ast snippet using the flux_semantic_analyzer_t and produce
// a semantic graph for that snippet.
struct flux_error_t *flux_analyze_with(struct flux_semantic_analyzer_t *, struct flux_ast_pkg_t *, struct flux_semantic_pkg_t **);

// flux_analyze analyzes the given AST and will populate the second pointer argument with
// a pointer to the resulting semantic graph.
// It is the caller's responsibility to free the resulting semantic graph with a call to flux_free_semantic_pkg().
// Note that the caller should free the pointer to the semantic graph, not the pointer to the pointer
// to the semantic graph.  It is the former that references memory allocated by Rust.
// If analysis fails, the second pointer argument wil be pointed at 0, and an error will be returned.
// Any non-null error must be freed by calling flux_free_error.
// Regardless of whether an error is returned, this function will consume and free its
// flux_ast_pkg_t* argument, so it should not be reused after calling this function.
struct flux_error_t *flux_analyze(struct flux_ast_pkg_t *, struct flux_semantic_pkg_t **);

// flux_free_semantic_pkg will release the memory associated with the given pointer.
void flux_free_semantic_pkg(struct flux_semantic_pkg_t*);

// flux_semantic_marshal_fb will marshal the given semantic graph as a flatbuffer into
// the given buffer. If successful, memory will be allocated for the data
// within the buffer and it is the caller's responsibility to free this
// data. If an error happens it will be returned. The error must be freed
// using flux_free_error if it is non-null.
struct flux_error_t *flux_semantic_marshal_fb(struct flux_semantic_pkg_t *, struct flux_buffer_t *);

#ifdef __cplusplus
}
#endif

#endif
