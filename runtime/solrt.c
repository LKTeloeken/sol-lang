#include <stdio.h>
#include <stdlib.h>
#include <string.h>

void sol_panic(const char *msg) {
    fprintf(stderr, "SOL flare: %s\n", msg);
    exit(1);
}

void sol_print(const char *msg) {
    if (msg != NULL) {
        printf("%s\n", msg);
    }
}

char *sol_i64_to_str(long long v) {
    char *buf = malloc(32);
    if (buf == NULL) {
        return NULL;
    }
    snprintf(buf, 32, "%lld", v);
    return buf;
}

char *sol_f64_to_str(double v) {
    char *buf = malloc(64);
    if (buf == NULL) {
        return NULL;
    }
    snprintf(buf, 64, "%g", v);
    return buf;
}

char *sol_bool_to_str(int v) {
    char *buf = malloc(8);
    if (buf == NULL) {
        return NULL;
    }
    strcpy(buf, v ? "true" : "false");
    return buf;
}

char *sol_concat(const char *a, const char *b) {
    if (a == NULL) {
        a = "";
    }
    if (b == NULL) {
        b = "";
    }
    size_t len = strlen(a) + strlen(b) + 1;
    char *out = malloc(len);
    if (out == NULL) {
        return NULL;
    }
    strcpy(out, a);
    strcat(out, b);
    return out;
}

typedef struct SolObject {
    char *class_name;
    struct SolObject *next;
    char *field_names[64];
    char *field_values[64];
    int field_count;
} SolObject;

static SolObject *sol_objects = NULL;

SolObject *sol_new(const char *class_name) {
    SolObject *obj = calloc(1, sizeof(SolObject));
    if (obj == NULL) {
        return NULL;
    }
    obj->class_name = strdup(class_name != NULL ? class_name : "Object");
    obj->next = sol_objects;
    sol_objects = obj;
    return obj;
}

void sol_set_field(SolObject *obj, const char *name, const char *value) {
    if (obj == NULL || name == NULL) {
        return;
    }
    for (int i = 0; i < obj->field_count; i++) {
        if (strcmp(obj->field_names[i], name) == 0) {
            free(obj->field_values[i]);
            obj->field_values[i] = value != NULL ? strdup(value) : strdup("");
            return;
        }
    }
    if (obj->field_count >= 64) {
        return;
    }
    obj->field_names[obj->field_count] = strdup(name);
    obj->field_values[obj->field_count] = value != NULL ? strdup(value) : strdup("");
    obj->field_count++;
}

const char *sol_get_field(SolObject *obj, const char *name) {
    if (obj == NULL || name == NULL) {
        return "";
    }
    for (int i = 0; i < obj->field_count; i++) {
        if (strcmp(obj->field_names[i], name) == 0) {
            return obj->field_values[i];
        }
    }
    return "";
}
