import { useState } from 'react'

function useTTLValue<T>(calculate: () => T, ttl = 1000): T {
  const [value, setValue] = useState<T>((null as unknown) as T)
  const [cached, setCached] = useState(false)

  if (cached) return value

  setValue(calculate())
  setCached(true)
  setTimeout(() => setCached(false), ttl)
  return value
}

export default useTTLValue
