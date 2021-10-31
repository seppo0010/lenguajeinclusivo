import React, { ReactElement } from 'react';
import Box from './Box';
import './Section.css';

interface SectionIface {
  title?: string | ReactElement
}
function Section(props: React.PropsWithChildren<SectionIface>) {
  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
  }
  return (
    <Box>
      <form onSubmit={handleSubmit}>
        <fieldset>
          {props.title && <legend>{props.title}</legend>}
          {props.children}
        </fieldset>
      </form>
    </Box>
  )
}

export default Section
