module.exports = {
  main: {
    fields: ['caratula', 'firmantes', 'titulo', 'content', 'URL'],
    storeFields: ['URL', 'ExpId', 'actId'],
    idField: 'numeroDeExpediente'
  },
  search: {
    prefix: term => term.length > 3,
    fuzzy: term => term.length > 3 ? 0.2 : false
  }
}
