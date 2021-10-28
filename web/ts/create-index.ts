const MiniSearch = require('minisearch');
import { promises as fs } from 'fs';
import * as path from 'path';
const { StringDecoder } = require('string_decoder');
const decoder = new StringDecoder('utf8');

const DATADIR = '../../data';

const run = async () => {
  // FIXME: duplicates Expediente.tsx
  const ms = new MiniSearch({
    fields: ['content', 'URL'],
    storeFields: ['content', 'URL', 'ExpId', 'actId'],
    idField: 'numeroDeExpediente'
  });

  const buf = await fs.readFile(path.join(DATADIR, 'a.json'));
  const { Actuaciones, numero, anio, cuij, ...expediente } = JSON.parse(decoder.write(buf));
  expediente.data = { numero, anio, cuij };

  for (let i = 0; i < Actuaciones.length; i++) {
    const { documentos, ...actuacion } = Actuaciones[i];
    for (let j = 0; j < documentos.length; j++) {
      let d = { ...expediente, ...actuacion, ...documentos[j] };
      ms.add(d);
    }
  }

  await fs.writeFile('index.json', JSON.stringify(ms.toJSON()))
}
run()
