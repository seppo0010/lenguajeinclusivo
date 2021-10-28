import React from 'react'
import './Button.css'

interface ButtonIface {
  onClick?: React.MouseEventHandler<HTMLAnchorElement>
  href?: string
  selected?: boolean
}

const Button = (props: React.PropsWithChildren<ButtonIface>) => {
  return (<a className={`button ${props.selected ? 'selected' : null}`} {...props}>
    {props.children}
  </a>)
}
export default Button;
