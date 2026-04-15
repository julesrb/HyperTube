import React from "react";

export default function Button({children, onClick, className, disabled=false}: {children: React.ReactNode, onClick: () => void, className?: string, disabled?: boolean}) {
    return (<button
        disabled={disabled}
        onClick={onClick}
        className={"text-nowrap px-5 h-10 text-white " + (disabled ? "bg-gray " : "bg-black hover:bg-black-hover ") + className}
    >{children}</button>);
}
