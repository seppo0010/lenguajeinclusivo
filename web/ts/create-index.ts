const MiniSearch = require('minisearch');
import { promises as fs } from 'fs';
import * as path from 'path';
import * as process from 'process';
const { StringDecoder } = require('string_decoder');
const decoder = new StringDecoder('utf8');

const DATADIR = '../../data';

const run = async () => {
  const ms = new MiniSearch({
    fields: ['content', 'URL'],
    storeFields: ['content', 'URL', 'ExpId'],
    idField: 'numeroDeExpediente'
  });

  const buf = await fs.readFile(process.argv[2]);
  const { Actuaciones, numero, anio, cuij, ...expediente } = JSON.parse(decoder.write(buf));
  expediente.data = { numero, anio, cuij };

  for (let i = 0; i < Actuaciones.length; i++) {
    const { documentos, ...actuacion } = Actuaciones[i];
    for (let j = 0; j < documentos.length; j++) {
      let d = { ...expediente, ...actuacion, ...documentos[j] };
      ms.add(d);
    }
  }

  await fs.writeFile(process.argv[3], JSON.stringify(ms.toJSON()))
}
run()
