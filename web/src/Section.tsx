import React from 'react';
import Box from './Box';
import './Section.css';

interface SectionIface {
  title?: string
}
function Section(props: React.PropsWithChildren<SectionIface>) {
  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
  }
  return (
    <Box id={props.title}>
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
