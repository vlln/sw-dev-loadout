import { readFileSync } from 'node:fs'

const BINARY = /\.(png|jpg|gif|zip|pdf)$/i

export function countFiles(paths) {
  const byExt = new Map()
  for (const path of paths) {
    if (BINARY.test(path)) continue
    const text = readFileSync(path, 'utf8')
    const lines = text.split('\n').length
    const ext = path.includes('.') ? path.slice(path.lastIndexOf('.')) : '(none)'
    byExt.set(ext, (byExt.get(ext) ?? 0) + lines)
  }
  return [...byExt.entries()].sort((a, b) => b[1] - a[1])
}
