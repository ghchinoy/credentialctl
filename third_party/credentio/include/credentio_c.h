// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

#ifndef CREDENTIO_CONTRIBUTIONS_NATIVE_CREDENTIO_C_H_
#define CREDENTIO_CONTRIBUTIONS_NATIVE_CREDENTIO_C_H_

#include <stddef.h>
#include <stdint.h>

#if defined(_WIN32)
#define CR_EXPORT __declspec(dllexport)
#elif defined(__GNUC__) || defined(__clang__)
#define CR_EXPORT __attribute__((visibility("default")))
#else
#define CR_EXPORT
#endif

#ifdef __cplusplus
extern "C" {
#endif

/// Status codes returned in `out_status`.
#define CR_STATUS_OK 0
#define CR_STATUS_NO_CREDENTIALS 1
#define CR_STATUS_INVALID_ARGUMENT 2
#define CR_STATUS_IO_ERROR 3
#define CR_STATUS_INTERNAL_ERROR 4

/// Opaque handle representing a Credentio validator instance.
typedef struct cr_validator cr_validator;

/// Creates a new Credentio validator.
///
/// @param claim_signer_trust_pem Optional PEM string with claim signer trust anchors (or NULL).
/// @param tsa_trust_pem Optional PEM string with TSA trust anchors (or NULL).
/// @param skip_trust_checks Non-zero to skip trust checks for local verification/testing.
/// @return Opaque validator pointer, or NULL on internal creation failure.
CR_EXPORT cr_validator* cr_validator_create(
    const char* claim_signer_trust_pem,
    const char* tsa_trust_pem,
    int skip_trust_checks);

/// Frees a validator instance.
CR_EXPORT void cr_validator_free(cr_validator* validator);

/// Validates a media asset file and returns its validation results as crJSON.
///
/// @param validator The validator instance.
/// @param file_path Full path to the asset file.
/// @param media_type Optional IANA media type (e.g. "image/jpeg"), or NULL to infer from extension.
/// @param out_status Pointer to an integer where status code (CR_STATUS_*) will be stored.
/// @return Malloc-allocated UTF-8 JSON string (crJSON format), or NULL on failure/no-credentials.
///         The caller must free the returned string using `cr_string_free()`.
CR_EXPORT char* cr_validate_file(
    cr_validator* validator,
    const char* file_path,
    const char* media_type,
    int* out_status);

/// Validates media asset bytes in memory and returns its validation results as crJSON.
///
/// @param validator The validator instance.
/// @param bytes Pointer to raw asset bytes.
/// @param count Length of bytes buffer in bytes.
/// @param media_type Optional IANA media type, or NULL to infer.
/// @param out_status Pointer to an integer where status code (CR_STATUS_*) will be stored.
/// @return Malloc-allocated UTF-8 JSON string (crJSON format), or NULL on failure/no-credentials.
///         The caller must free the returned string using `cr_string_free()`.
CR_EXPORT char* cr_validate_bytes(
    cr_validator* validator,
    const uint8_t* bytes,
    size_t count,
    const char* media_type,
    int* out_status);

/// Returns the last error message recorded on the validator instance, or empty string.
/// The returned string pointer is valid until the next call on this validator.
CR_EXPORT const char* cr_last_error(cr_validator* validator);

/// Returns the engine-internal validation time (excluding file I/O / bridge overhead) in seconds.
CR_EXPORT double cr_last_internal_seconds(cr_validator* validator);

/// Frees a string previously returned by `cr_validate_file` or `cr_validate_bytes`.
CR_EXPORT void cr_string_free(char* str);

/// Returns the version of the Credentio C-ABI wrapper.
CR_EXPORT const char* cr_version(void);

#ifdef __cplusplus
}
#endif

#endif  // CREDENTIO_CONTRIBUTIONS_NATIVE_CREDENTIO_C_H_
