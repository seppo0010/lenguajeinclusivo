import React from 'react';
import Box from './Box';
import Button from './Button';

export interface DocumentData {
  URL: string;
  content: string;
}

function Document(props: React.PropsWithChildren<DocumentData>) {
  return (<Box>
    <div>
      {props.content}
    </div>
    <div>
      <Button>expand</Button>
      <Button href={props.URL}>download</Button>
    </div>
  </Box>)
}
export default Document
