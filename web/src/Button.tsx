import React from 'react'
import './Button.css'

interface ButtonIface {
  onClick?: React.MouseEventHandler<HTMLAnchorElement>
  selected?: boolean
}

const Button = (props: React.PropsWithChildren<ButtonIface>) => {
  return (<a className={`button ${props.selected ? 'selected' : null}`} onClick={props.onClick}>
    {props.children}
  </a>)
}
export default Button;
