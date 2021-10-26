const MiniSearch = require('minisearch');
import { promises as fs } from 'fs';
import * as path from 'path';
const { StringDecoder } = require('string_decoder');
const decoder = new StringDecoder('utf8');

const DATADIR = '../data';

const JSONSlice = (buf, s, e) => {
  const json = decoder.write(buf.slice(s, e))
  return JSON.parse(json)
}
const importJSONArray = async (ms, f) => {
  const buf = await fs.readFile(f)
  let j = 0;
  let json = '';
  for (var i = 0; i < buf.length; ++i) {
    if (buf[i] === '\n'.charCodeAt(0)) {
      ms.add(JSONSlice(buf, j, i)._source)
      j = i;
    }
  }
  if (i - j > 10) {
    ms.add(JSONSlice(buf, j, i)._source)
  }
}

const run = async () => {
  const ms = new MiniSearch({
    fields: ['text', 'URL'],
    storeFields: ['text', 'URL', 'actuacionId'],
    idField: 'numeroDeExpediente'
  });

  await importJSONArray(ms, path.join(DATADIR, 'document.json'))
  await fs.writeFile('index.json', JSON.stringify(ms.toJSON()))
}
run()
