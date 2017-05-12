#ifndef LIBJSONNET_BRIDGE_H
#define LIBJSONNET_BRIDGE_H
#include <libjsonnet.h>

typedef JsonnetImportCallback* JsonnetImportCallbackPtr;

struct JsonnetVm* go_get_guts(void* ctx);

char* CallImport_cgo(void *ctx, const char *base, const char *rel, char **found_here, int *success);

char* go_call_import(void* vm, char *base, char *rel, char **path, int *success);

#endif
