import React from "react";
import {CrossIcon} from "@/components/Icon";

export function Button({children, onClick, className, disabled=false}: {children: React.ReactNode, onClick: () => void, className?: string, disabled?: boolean}) {
    return (<button
        disabled={disabled}
        onClick={onClick}
        className={"text-nowrap px-5 h-10 text-white " + (disabled ? "bg-gray " : "bg-black hover:bg-black-hover ") + className}
    >{children}</button>);
}

export function CloseButton({onClick, className, disabled=false}: {onClick: () => void, className?: string, disabled?: boolean}) {
    const [isHover, setIsHover] = React.useState(false);
    let color;

    if (disabled)
        color = "white";
    else if (isHover)
        color = "black-hover";
    else
        color = "black";

    return (<button disabled={disabled} onClick={onClick}
        onMouseEnter={() => setIsHover(true)}
        onMouseLeave={() => setIsHover(false)}
        className={"text-nowrap " + className}
    ><CrossIcon color={color} className={(isHover ? "stroke-2 stroke-black-hover" : "")}/></button>);
}
