import React from 'react'
import './Button.css'

interface ButtonIface {
  onClick?: Function
  href?: string
  selected?: boolean
}

const Button = ({ selected, onClick, ...props }: React.PropsWithChildren<ButtonIface>) => {
  const handleOnClick = (event: React.MouseEvent<HTMLAnchorElement>) => {
    onClick && onClick(props)
  }

  return (<a className={`button ${selected ? 'selected' : null}`}
    onClick={handleOnClick} {...props}>
    {props.children}
  </a>)
}
export default Button;
