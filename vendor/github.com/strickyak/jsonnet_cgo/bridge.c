#include <memory.h>
#include <stdio.h>
#include <string.h>
#include <libjsonnet.h>
#include "bridge.h"

char* CallImport_cgo(void *ctx, const char *base, const char *rel, char **found_here, int *success) {
  struct JsonnetVm* vm = ctx;
  char *path = NULL;
  char* result = go_call_import(vm, (char*)base, (char*)rel, &path, success);
  if (*success) {
    char* found_here_buf = jsonnet_realloc(vm, NULL, strlen(path)+1);
    strcpy(found_here_buf, path);
    *found_here = found_here_buf;
  }
  char* buf = jsonnet_realloc(vm, NULL, strlen(result)+1);
  strcpy(buf, result);
  return buf;
}
