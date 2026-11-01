import { countFiles } from './count.mjs'

const args = process.argv.slice(2)
if (args.length === 0) {
  console.error('usage: tally <file...>')
  process.exit(2)
}

const rows = countFiles(args)
const width = Math.max(...rows.map(([ext]) => ext.length))
for (const [ext, lines] of rows) {
  console.log(`${ext.padEnd(width)}  ${lines}`)
}
