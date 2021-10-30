import React, { useState } from 'react';
import Box from './Box';
import Button from './Button';

export interface DocumentData {
  URL: string;
  content: string;
}

function Document(props: React.PropsWithChildren<DocumentData>) {
  const [expanded, setExpanded] = useState(false);

  return (<Box>
    <div>
      {props.content.slice(0, expanded ? undefined : 200)}
    </div>
    <div>
      <Button onClick={() => setExpanded(!expanded)} selected={expanded}>expand</Button>
      <Button href={props.URL}>download</Button>
    </div>
  </Box >)
}
export default Document
