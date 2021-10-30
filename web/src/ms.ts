const MiniSearch = require('minisearch');
export const MiniSearchConfig = {
  fields: ['content', 'URL'],
  storeFields: ['content', 'URL', 'ExpId', 'actId'],
  idField: 'numeroDeExpediente'
}
const ms = () => new MiniSearch(MiniSearchConfig)

export default ms;
