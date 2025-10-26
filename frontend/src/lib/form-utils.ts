import type { FieldValues, UseFormSetError } from 'react-hook-form'

/**
 * Extracts field names from API error response
 */
export function extractErrorFields(error: unknown): Record<string, string> {
  if (!error || typeof error !== 'object') return {}

  const errorObj = error as Record<string, unknown>

  // Handle validation errors with field mapping
  if (errorObj.errors && typeof errorObj.errors === 'object') {
    const errors = errorObj.errors as Record<string, unknown>
    return Object.entries(errors).reduce(
      (acc, [key, value]) => {
        acc[key] = typeof value === 'string' ? value : String(value)
        return acc
      },
      {} as Record<string, string>
    )
  }

  return {}
}

/**
 * Sets form field errors from API response
 */
export function setFormErrors<T extends FieldValues>(
  error: unknown,
  setError: UseFormSetError<T>,
  defaultMessage = 'An error occurred'
): void {
  const fieldErrors = extractErrorFields(error)

  if (Object.keys(fieldErrors).length > 0) {
    Object.entries(fieldErrors).forEach(([field, message]) => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      setError(field as any, {
        type: 'server',
        message,
      })
    })
  } else {
    // Set root error if no field-specific errors
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    setError('root' as any, {
      type: 'server',
      message: defaultMessage,
    })
  }
}

/**
 * Get error message from API response
 */
export function getErrorMessage(error: unknown): string {
  if (!error || typeof error !== 'object') {
    return 'An unexpected error occurred'
  }

  const errorObj = error as Record<string, unknown>

  if (typeof errorObj.message === 'string') {
    return errorObj.message
  }

  if (typeof errorObj.detail === 'string') {
    return errorObj.detail
  }

  if (errorObj.status && errorObj.statusCode) {
    const status = errorObj.statusCode as number
    if (status === 401) return 'Unauthorized. Please log in again.'
    if (status === 403) return 'Forbidden. You do not have permission.'
    if (status === 404) return 'Resource not found.'
    if (status === 422) return 'Invalid input. Please check your data.'
    if (status >= 500) return 'Server error. Please try again later.'
  }

  return 'An error occurred. Please try again.'
}
