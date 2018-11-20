/* Code generated by cmd/cgo; DO NOT EDIT. */

/* package _/d_/work/Filter-Windows/GoProxyDotNet/goproxy */


#line 1 "cgo-builtin-prolog"

#include <stddef.h> /* for ptrdiff_t below */

#ifndef GO_CGO_EXPORT_PROLOGUE_H
#define GO_CGO_EXPORT_PROLOGUE_H

typedef struct { const char *p; ptrdiff_t n; } _GoString_;

#endif

/* Start of preamble from import "C" comments.  */


#line 3 "main.go"

typedef void (*callback)(long long id);

static inline void FireCallback(void *ptr, long long id)
{
	callback p = (callback)ptr;
	p(id);
}


#line 1 "cgo-generated-wrapper"




/* End of preamble from import "C" comments.  */


/* Start of boilerplate cgo prologue.  */
#line 1 "cgo-gcc-export-header-prolog"

#ifndef GO_CGO_PROLOGUE_H
#define GO_CGO_PROLOGUE_H

typedef signed char GoInt8;
typedef unsigned char GoUint8;
typedef short GoInt16;
typedef unsigned short GoUint16;
typedef int GoInt32;
typedef unsigned int GoUint32;
typedef long long GoInt64;
typedef unsigned long long GoUint64;
typedef GoInt64 GoInt;
typedef GoUint64 GoUint;
typedef __SIZE_TYPE__ GoUintptr;
typedef float GoFloat32;
typedef double GoFloat64;
typedef float _Complex GoComplex64;
typedef double _Complex GoComplex128;

/*
  static assertion to make sure the file is being used on architecture
  at least with matching size of GoInt.
*/
typedef char _check_for_64_bit_pointer_matching_GoInt[sizeof(void*)==64/8 ? 1:-1];

typedef _GoString_ GoString;
typedef void *GoMap;
typedef void *GoChan;
typedef struct { void *t; void *v; } GoInterface;
typedef struct { void *data; GoInt len; GoInt cap; } GoSlice;

#endif

/* End of boilerplate cgo prologue.  */

#ifdef __cplusplus
extern "C" {
#endif


extern void SetOnBeforeRequestCallback(void* p0);

extern void SetOnBeforeResponseCallback(void* p0);

extern void Init(GoInt16 p0);

extern void Start();

extern void Stop();

extern GoUint8 IsRunning();

extern void GetCert(GoSlice* p0);

extern GoUint8 RequestGetUrl(GoInt64 p0, GoString* p1);

extern GoUint8 RequestGetBody(GoInt64 p0, GoSlice* p1);

extern GoUint8 RequestGetBodyAsString(GoInt64 p0, GoString* p1);

extern GoUint8 RequestHasBody(GoInt64 p0);

extern GoUint8 RequestHeaderExists(GoInt64 p0, GoString p1);

extern GoUint8 RequestGetFirstHeader(GoInt64 p0, GoString p1, GoString* p2);

extern GoUint8 RequestSetHeader(GoInt64 p0, GoString p1, GoString p2);

extern GoInt ResponseGetStatusCode(GoInt64 p0);

extern GoUint8 ResponseGetBody(GoInt64 p0, GoSlice* p1);

extern GoUint8 ResponseGetBodyAsString(GoInt64 p0, GoString* p1);

extern GoUint8 ResponseHasBody(GoInt64 p0);

extern GoUint8 ResponseHeaderExists(GoInt64 p0, GoString p1);

extern GoUint8 ResponseGetFirstHeader(GoInt64 p0, GoString p1, GoString* p2);

extern GoUint8 ResponseSetHeader(GoInt64 p0, GoString p1, GoString p2);

extern GoUint8 CreateResponse(GoInt64 p0, GoInt p1, GoString p2, GoString p3);

#ifdef __cplusplus
}
#endif
