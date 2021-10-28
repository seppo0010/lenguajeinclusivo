import React from 'react';
import './Box.css';

interface BoxIface {
  id?: string
}

function Box(props: React.PropsWithChildren<BoxIface>) {
  return (
    <div className="box" id={props.id}>
      {props.children}
    </div>
  )
}

export default Box;
