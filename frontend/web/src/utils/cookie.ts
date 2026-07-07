export function getCookie(name: string): string {
  const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'))
  return match ? decodeURIComponent(match[2]) : ''
}

export function setCookie(name: string, value: string, expiryDays = 1) {
  const expires = new Date()
  expires.setDate(expires.getDate() + expiryDays)
  document.cookie = `${name}=${encodeURIComponent(value)};expires=${expires.toUTCString()};path=/`
}
