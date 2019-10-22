#ifndef _INFLUXDATA_FLUX_H
#define _INFLUXDATA_FLUX_H

#ifdef __cplusplus
extern "C" {
#endif

// flux_ast_t is the AST representation of a flux query.
struct flux_ast_t;

// flux_parse will take in a string and return the AST representation
// of the query.
struct flux_ast_t *flux_parse(const char *);

// flux_ast_free will free memory associated with an AST handle.
void flux_ast_free(struct flux_ast_t *);

#ifdef __cplusplus
}
#endif

#endif
