
#include "influxdata/flux.h"

#include <assert.h>
#include <stdio.h>

void test_ast();
void test_semantic();

int main(int argc, char* argv[]) {
  test_ast();
  test_semantic();
  return 0;
}

void test_ast() {
  printf("Testing AST functions...\n");

  printf("Parsing to AST\n");
  struct flux_ast_pkg_t *ast_pkg_foo = flux_parse("package foo\nx = 1 + 1");
  assert(ast_pkg_foo !=  NULL);

  printf("Marshaling to JSON\n");
  struct flux_buffer_t buf;
  // it's unclear how to test errors returned by serialization
  struct flux_error_t *err = flux_ast_marshal_json(ast_pkg_foo, &buf);
  assert(err == NULL);
  printf("  json buffer is length %ld\n", buf.len);
  flux_free_bytes(buf.data);

  printf("Marshaling to FlatBuffer\n");
  err = flux_ast_marshal_fb(ast_pkg_foo, &buf);
  assert(err == NULL);
  printf("  FlatBuffer is length %ld\n", buf.len);
  flux_free_bytes(buf.data);

  flux_free_ast_pkg(ast_pkg_foo);
  printf("\n");
}

void test_semantic() {
  printf("Testing semantic graph functions...\n");

  {
    printf("Parsing to AST\n");
    struct flux_ast_pkg_t *ast_pkg_foo = flux_parse("package foo\nx = 1 + 1");
    assert(ast_pkg_foo !=  NULL);

    printf("Analyzing (expect success)\n");
    struct flux_semantic_pkg_t* sem_pkg = NULL;
    struct flux_error_t* err = flux_analyze(ast_pkg_foo, &sem_pkg);
    assert(err == NULL);

    printf("Marshaling to FlatBuffer\n");
    struct flux_buffer_t buf;
    err = flux_semantic_marshal_fb(sem_pkg, &buf);
    assert(err == NULL);
    printf("  FlatBuffer is length %ld\n", buf.len);
    flux_free_bytes(buf.data);

    flux_free_semantic_pkg(sem_pkg);
  }

  {
    printf("Parsing to AST\n");
    struct flux_ast_pkg_t *ast_pkg_foo = flux_parse("package foo\nx = 1 + 1.0");
    assert(ast_pkg_foo !=  NULL);

    printf("Analyzing (expect failure)\n");
    struct flux_semantic_pkg_t* sem_pkg = NULL;
    struct flux_error_t* err = flux_analyze(ast_pkg_foo, &sem_pkg);
    assert(err != NULL);
    assert(sem_pkg == NULL);
    const char* err_str = flux_error_str(err);
    printf("  error: %s\n", err_str);
    flux_free_bytes(err_str);
    flux_free_error(err);
  }

  printf("\n");
}
