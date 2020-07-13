
#include "influxdata/flux.h"

#include <assert.h>
#include <stdio.h>

void test_ast();
void test_semantic();
void test_semantic_analyzer();
void test_env_stdlib();

int main(int argc, char* argv[]) {
  test_ast();
  test_semantic();
  test_semantic_analyzer();
  test_env_stdlib();
  return 0;
}

void test_ast() {
  printf("Testing AST functions...\n");

  {
    printf("Parsing to AST (expect success)\n");
    struct flux_ast_pkg_t *ast_pkg_foo = flux_parse("test", "package foo\nx = 1 + 1");
    assert(ast_pkg_foo !=  NULL);

    struct flux_error_t* err = flux_ast_get_error(ast_pkg_foo);
    assert(err == NULL);

    printf("Marshaling to JSON\n");
    struct flux_buffer_t buf;
    // it's unclear how to test errors returned by serialization
    err = flux_ast_marshal_json(ast_pkg_foo, &buf);
    assert(err == NULL);
    printf("  json buffer is length %ld\n", buf.len);
    flux_free_bytes(buf.data);

    printf("Marshaling to FlatBuffer\n");
    err = flux_ast_marshal_fb(ast_pkg_foo, &buf);
    assert(err == NULL);
    printf("  FlatBuffer is length %ld\n", buf.len);
    flux_free_bytes(buf.data);

    flux_free_ast_pkg(ast_pkg_foo);
  }
  {
    printf("Parsing to AST (expect failure)\n");
    struct flux_ast_pkg_t *ast_pkg_foo = flux_parse("test", "x = 1 + / 1");
    assert(ast_pkg_foo !=  NULL);

    struct flux_error_t* err = flux_ast_get_error(ast_pkg_foo);
    assert(err != NULL);
    const char* err_str = flux_error_str(err);
    printf("  error: %s\n", err_str);
    flux_free_bytes(err_str);
    flux_free_error(err);
    flux_free_ast_pkg(ast_pkg_foo);
    printf("\n");
  }
}

void test_semantic() {
  printf("Testing semantic graph functions...\n");

  {
    printf("Parsing to AST\n");
    struct flux_ast_pkg_t *ast_pkg_foo = flux_parse("test", "package foo\nx = 1 + 1");
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
    struct flux_ast_pkg_t *ast_pkg_foo = flux_parse("test", "package foo\nx = 1 + 1.0");
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

  {
    printf("Parsing to AST\n");
    struct flux_ast_pkg_t *ast_pkg_foo = flux_parse("test", "package foo\nx = 1 + 1");
    assert(ast_pkg_foo !=  NULL);
    printf("Find variable type v (expect success)\n");
    struct flux_buffer_t buf;
    struct flux_error_t* err = flux_find_var_type(ast_pkg_foo, "v", &buf);
    assert(err == NULL);
    printf("  FlatBuffer is length %ld\n", buf.len);
    flux_free_bytes(buf.data);
  }

  {
    printf("Parsing to AST\n");
    struct flux_ast_pkg_t *ast_pkg_foo = flux_parse("test", "package foo\nx = 1 + 1.0");
    assert(ast_pkg_foo !=  NULL);
    printf("Find variable type v (expect failure)\n");
    struct flux_buffer_t buf;
    struct flux_error_t* err = flux_find_var_type(ast_pkg_foo, "v", &buf);
    assert(err != NULL);
    const char* err_str = flux_error_str(err);
    printf("  error: %s\n", err_str);
    flux_free_bytes(err_str);
    flux_free_error(err);
  }

  printf("\n");
}

void test_semantic_analyzer() {
  printf("Testing semantic analyzer...\n");

  struct flux_semantic_analyzer_t *analyzer = flux_new_semantic_analyzer("main");

  struct flux_ast_pkg_t *ast_pkg = NULL;
  struct flux_semantic_pkg_t *sem_pkg = NULL;
  struct flux_error_t *err = NULL;

  printf("Parsing and analyzing \"x = 10\"\n");
  ast_pkg = flux_parse("test", "x = 10");
  assert(ast_pkg != NULL);
  err = flux_analyze_with(analyzer, ast_pkg, &sem_pkg);
  assert(err == NULL);
  assert(sem_pkg != NULL);
  flux_free_semantic_pkg(sem_pkg);

  printf("Parsing and analyzing \"y = x * x\"\n");
  ast_pkg = flux_parse("test", "y = x * x");
  assert(ast_pkg != NULL);
  sem_pkg = NULL;
  err = flux_analyze_with(analyzer, ast_pkg, &sem_pkg);
  assert(err == NULL);
  assert(sem_pkg != NULL);
  flux_free_semantic_pkg(sem_pkg);

  printf("Parsing and analyzing \"z = a + y\" (expect failure)\n");
  ast_pkg = flux_parse("test", "z = a + y");
  assert(ast_pkg != NULL);
  sem_pkg = NULL;
  err = flux_analyze_with(analyzer, ast_pkg, &sem_pkg);
  assert(err != NULL);
  assert(sem_pkg == NULL);
  const char* err_str = flux_error_str(err);
  printf("  error: %s\n", err_str);
  flux_free_bytes(err_str);
  flux_free_error(err);

  flux_free_semantic_analyzer(analyzer);
  printf("\n");
}

void test_env_stdlib() {
  printf("Testing flux_get_env_stdlib\n");
  struct flux_buffer_t buf;
  flux_get_env_stdlib(&buf);
  assert(buf.data != NULL);
  printf("  got a buffer of size %ld\n", buf.len);
  flux_free_bytes(buf.data);
  printf("\n");
}
