import React, { useState } from 'react';
import Box from './Box';
import Button from './Button';
import { SearchData } from './Search';
import Highlighter from './Highlighter';

export interface DocumentData {
  URL: string;
  content: string;
}

export interface DocumentProps {
  data: DocumentData
  search: SearchData
}
function Document({ data, search }: React.PropsWithChildren<DocumentProps>) {
  const [expanded, setExpanded] = useState(false);

  return (<Box>
    <div>
      <Highlighter text={data.content} term={search.term} length={expanded ? 0 : 200} />
    </div>
    <div>
      <Button onClick={() => setExpanded(!expanded)} selected={expanded}>expand</Button>
      <Button href={data.URL}>download</Button>
    </div>
  </Box >)
}
export default Document
