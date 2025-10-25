/* eslint-disable @typescript-eslint/no-explicit-any */
/**
 * Convert string from camelCase to snake_case
 */
export function camelToSnake(str: string): string {
  return str.replace(/[A-Z]/g, (letter) => `_${letter.toLowerCase()}`)
}

/**
 * Convert string from snake_case to camelCase
 */
export function snakeToCamel(str: string): string {
  return str.replace(/_([a-z])/g, (_, letter) => letter.toUpperCase())
}

/**
 * Convert object keys from camelCase to snake_case recursively
 */
export function convertKeysToSnake(obj: any): any {
  if (obj === null || obj === undefined) {
    return obj
  }

  if (Array.isArray(obj)) {
    return obj.map((item) => convertKeysToSnake(item))
  }

  if (typeof obj === 'object' && obj.constructor === Object) {
    const converted: any = {}
    for (const key in obj) {
      if (Object.prototype.hasOwnProperty.call(obj, key)) {
        const snakeKey = camelToSnake(key)
        converted[snakeKey] = convertKeysToSnake(obj[key])
      }
    }
    return converted
  }

  return obj
}

/**
 * Convert object keys from snake_case to camelCase recursively
 */
export function convertKeysToCamel(obj: any): any {
  if (obj === null || obj === undefined) {
    return obj
  }

  if (Array.isArray(obj)) {
    return obj.map((item) => convertKeysToCamel(item))
  }

  if (typeof obj === 'object' && obj.constructor === Object) {
    const converted: any = {}
    for (const key in obj) {
      if (Object.prototype.hasOwnProperty.call(obj, key)) {
        const camelKey = snakeToCamel(key)
        converted[camelKey] = convertKeysToCamel(obj[key])
      }
    }
    return converted
  }

  return obj
}
