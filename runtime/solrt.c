#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>
#include <math.h>
#include <sys/stat.h>
#include <ctype.h>

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

/* --- script args (skip argv[0] = program name) --- */
static int sol_argc = 0;
static char **sol_argv = NULL;

void sol_args_init(int argc, char **argv) {
    sol_argc = argc;
    sol_argv = argv;
    srand((unsigned int)time(NULL));
}

long long sol_args_count(void) {
    if (sol_argc <= 1) {
        return 0;
    }
    return (long long)(sol_argc - 1);
}

char *sol_args_at(long long i) {
    if (i < 0 || sol_argv == NULL || (int)(i + 1) >= sol_argc) {
        sol_panic("Args.at index out of range");
    }
    return strdup(sol_argv[(int)i + 1]);
}

/* --- Time --- */
long long sol_time_now(void) {
    return (long long)time(NULL);
}

void sol_sleep_ms(long long ms) {
    if (ms > 0) {
        usleep((useconds_t)(ms * 1000));
    }
}

char *sol_time_format(long long unix_sec, const char *layout) {
    if (layout == NULL) {
        layout = "2006-01-02 15:04:05";
    }
    time_t t = (time_t)unix_sec;
    struct tm *tm_info = localtime(&t);
    if (tm_info == NULL) {
        return strdup("");
    }
    char buf[256];
    if (strftime(buf, sizeof(buf), layout, tm_info) == 0) {
        return strdup("");
    }
    return strdup(buf);
}

/* --- String (no split in native backend) --- */
long long sol_str_len(const char *s) {
    return (long long)strlen(s != NULL ? s : "");
}

char *sol_str_trim(const char *s) {
    if (s == NULL) {
        return strdup("");
    }
    size_t n = strlen(s);
    size_t start = 0;
    while (start < n && isspace((unsigned char)s[start])) {
        start++;
    }
    size_t end = n;
    while (end > start && isspace((unsigned char)s[end - 1])) {
        end--;
    }
    size_t len = end - start;
    char *out = malloc(len + 1);
    if (out == NULL) {
        return NULL;
    }
    memcpy(out, s + start, len);
    out[len] = '\0';
    return out;
}

int sol_str_contains(const char *s, const char *sub) {
    if (s == NULL || sub == NULL) {
        return 0;
    }
    return strstr(s, sub) != NULL ? 1 : 0;
}

char *sol_str_substring(const char *s, long long start, long long end) {
    if (s == NULL) {
        return strdup("");
    }
    size_t n = strlen(s);
    if (start < 0 || end < start || (size_t)end > n) {
        sol_panic("substring: invalid range");
    }
    size_t len = (size_t)(end - start);
    char *out = malloc(len + 1);
    if (out == NULL) {
        return NULL;
    }
    memcpy(out, s + start, len);
    out[len] = '\0';
    return out;
}

/* --- Math --- */
double sol_math_abs(double x) {
    return fabs(x);
}

double sol_math_min(double a, double b) {
    return fmin(a, b);
}

double sol_math_max(double a, double b) {
    return fmax(a, b);
}

long long sol_math_floor(double x) {
    return (long long)floor(x);
}

double sol_math_random(void) {
    return (double)rand() / ((double)RAND_MAX + 1.0);
}

/* --- File --- */
char *sol_file_read(const char *path) {
    if (path == NULL) {
        sol_panic("fileRead: null path");
    }
    FILE *f = fopen(path, "rb");
    if (f == NULL) {
        sol_panic("fileRead: cannot open file");
    }
    fseek(f, 0, SEEK_END);
    long sz = ftell(f);
    fseek(f, 0, SEEK_SET);
    if (sz < 0) {
        fclose(f);
        sol_panic("fileRead: cannot read file");
    }
    char *buf = malloc((size_t)sz + 1);
    if (buf == NULL) {
        fclose(f);
        return NULL;
    }
    size_t n = fread(buf, 1, (size_t)sz, f);
    fclose(f);
    buf[n] = '\0';
    return buf;
}

void sol_file_write(const char *path, const char *content) {
    if (path == NULL) {
        sol_panic("fileWrite: null path");
    }
    FILE *f = fopen(path, "wb");
    if (f == NULL) {
        sol_panic("fileWrite: cannot open file");
    }
    if (content != NULL) {
        fputs(content, f);
    }
    fclose(f);
}

void sol_file_append(const char *path, const char *content) {
    if (path == NULL) {
        sol_panic("fileAppend: null path");
    }
    FILE *f = fopen(path, "ab");
    if (f == NULL) {
        sol_panic("fileAppend: cannot open file");
    }
    if (content != NULL) {
        fputs(content, f);
    }
    fclose(f);
}

int sol_file_exists(const char *path) {
    if (path == NULL) {
        return 0;
    }
    struct stat st;
    return stat(path, &st) == 0 ? 1 : 0;
}
