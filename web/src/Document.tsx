import React from 'react';
export interface DocumentData {
  URL: string
  content: string
}

function Document(props: React.Props<DocumentData>) {
  return (<p>document</p>)
}
export default Document
