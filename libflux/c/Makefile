VALGRIND = valgrind
VALGRIND_ARGS = --leak-check=full --error-exitcode=1

BIN_STATIC = libflux_memory_tester_static
BIN_DYNAMIC = libflux_memory_tester_dynamic

SOURCES = *.c

TARGET_DIR = $(PWD)/../target
INCLUDE_DIR = $(PWD)/../include

test-valgrind: $(BIN_STATIC) $(BIN_DYNAMIC)
	$(VALGRIND) $(VALGRIND_ARGS) ./$(BIN_STATIC)
	LD_LIBRARY_PATH=$(TARGET_DIR)/debug $(VALGRIND) $(VALGRIND_ARGS) ./$(BIN_DYNAMIC)

$(BIN_STATIC): $(SOURCES)
	$(CC) -g -Wall -Werror $(SOURCES) -I$(INCLUDE_DIR) \
		$(TARGET_DIR)/debug/libflux.a \
		-o $@ -lpthread -ldl

$(BIN_DYNAMIC): $(SOURCES)
	$(CC) -g -Wall -Werror $(SOURCES) -I$(INCLUDE_DIR) \
		-L $(TARGET_DIR)/debug -lflux \
		-o $@

clean:
	rm -f $(BIN_STATIC) $(BIN_DYNAMIC)

.PHONY: test-valgrind clean
