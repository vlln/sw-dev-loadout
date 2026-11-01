import { test } from 'node:test'
import assert from 'node:assert/strict'
import { writeFileSync, mkdtempSync } from 'node:fs'
import { join } from 'node:path'
import { tmpdir } from 'node:os'
import { countFiles } from '../src/count.mjs'

test('counts lines per extension', () => {
  const dir = mkdtempSync(join(tmpdir(), 'tally-'))
  const a = join(dir, 'a.ts'); writeFileSync(a, 'one\ntwo\n')
  const b = join(dir, 'b.md'); writeFileSync(b, 'x\n')
  const rows = countFiles([a, b])
  assert.deepEqual(rows, [['.ts', 3], ['.md', 2]])
})
