
#include "influxdata/flux.h"

#include <assert.h>
#include <stdio.h>

void test_ast();
void test_semantic();
void test_stateful_analyzer();
void test_env_stdlib();

int main(int argc, char* argv[]) {
  test_ast();
  test_semantic();
  test_stateful_analyzer();
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
    flux_free_error(err);
    flux_free_ast_pkg(ast_pkg_foo);
    printf("\n");
  }
  {
    printf("Format AST\n");
    struct flux_ast_pkg_t *ast_pkg_foo = flux_parse("test", "package foo\nx=1+1");
    assert(ast_pkg_foo != NULL);

    struct flux_error_t* err = flux_ast_get_error(ast_pkg_foo);
    assert(err == NULL);

    struct flux_buffer_t buf;
    err = flux_ast_format(ast_pkg_foo, &buf);
    assert(err == NULL);

    flux_free_ast_pkg(ast_pkg_foo);
    flux_free_bytes(buf.data);
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
    struct flux_error_t* err = flux_analyze(ast_pkg_foo, "", &sem_pkg);
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
    struct flux_error_t* err = flux_analyze(ast_pkg_foo, "", &sem_pkg);
    assert(err != NULL);
    assert(sem_pkg != NULL);
    const char* err_str = flux_error_str(err);
    printf("  error: %s\n", err_str);
    flux_free_error(err);
    flux_free_semantic_pkg(sem_pkg);
  }

  {
    printf("Parsing to AST\n");
    struct flux_ast_pkg_t *ast_pkg_foo = flux_parse("test", "package foo\nx = 1 + 1");
    assert(ast_pkg_foo != NULL);

    struct flux_semantic_pkg_t* sem_pkg = NULL;
    struct flux_error_t* err = flux_analyze(ast_pkg_foo, "", &sem_pkg);
    assert(err == NULL);
    assert(sem_pkg != NULL);

    printf("Find variable type v (expect success)\n");
    struct flux_buffer_t buf;
    err = flux_find_var_type(sem_pkg, "v", &buf);
    // Note that we do not call flux_free_ast_pkg(ast_pkg_foo); here because we will
    // consume the AST package during the conversion from the AST package to the semantic package.
    assert(err == NULL);
    printf("  FlatBuffer is length %ld\n", buf.len);
    flux_free_bytes(buf.data);
    flux_free_semantic_pkg(sem_pkg);
  }

  {
    printf("Parsing to AST\n");
    struct flux_ast_pkg_t *ast_pkg_foo = flux_parse("test", "package foo\nx = 1 + 1.0");
    assert(ast_pkg_foo !=  NULL);

    struct flux_semantic_pkg_t* sem_pkg = NULL;
    struct flux_error_t* err = flux_analyze(ast_pkg_foo, "", &sem_pkg);
    assert(err != NULL);
    assert(sem_pkg != NULL);
    const char* err_str = flux_error_str(err);
    printf("  error: %s\n", err_str);
    flux_free_error(err);

    printf("Find variable type v (expect failure)\n");
    struct flux_buffer_t buf;
    err = flux_find_var_type(sem_pkg, "v", &buf);
    assert(err == NULL);

    flux_free_bytes(buf.data);
    flux_free_semantic_pkg(sem_pkg);
  }

  printf("\n");
}

void test_stateful_analyzer() {
  printf("Testing semantic analyzer...\n");

  struct flux_stateful_analyzer_t *analyzer = flux_new_stateful_analyzer("");

  struct flux_ast_pkg_t *ast_pkg = NULL;
  struct flux_semantic_pkg_t *sem_pkg = NULL;
  struct flux_error_t *err = NULL;

  printf("Parsing and analyzing \"x = 10\"\n");
  const char* src = "x = 10";
  ast_pkg = flux_parse("test", src);
  assert(ast_pkg != NULL);
  err = flux_analyze_with(analyzer, src, ast_pkg, &sem_pkg);
  assert(err == NULL);
  assert(sem_pkg != NULL);
  flux_free_semantic_pkg(sem_pkg);

  printf("Parsing and analyzing \"y = x * x\"\n");
  ast_pkg = flux_parse("test", "y = x * x");
  assert(ast_pkg != NULL);
  sem_pkg = NULL;
  err = flux_analyze_with(analyzer, NULL, ast_pkg, &sem_pkg);
  assert(err == NULL);
  assert(sem_pkg != NULL);
  flux_free_semantic_pkg(sem_pkg);

  printf("Parsing and analyzing \"z = a + y\" (expect failure)\n");
  ast_pkg = flux_parse("test", "z = a + y");
  assert(ast_pkg != NULL);
  sem_pkg = NULL;
  err = flux_analyze_with(analyzer, NULL, ast_pkg, &sem_pkg);
  assert(err != NULL);
  assert(sem_pkg == NULL);
  const char* err_str = flux_error_str(err);
  printf("  error: %s\n", err_str);
  flux_error_print(err);
  flux_free_error(err);

  flux_free_stateful_analyzer(analyzer);
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
