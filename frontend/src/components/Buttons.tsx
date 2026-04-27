import React from "react";
import {CrossIcon} from "@/components/Icons";

export function Button({children, onClick, className, disabled=false}: {children: React.ReactNode, onClick: () => void, className?: string, disabled?: boolean}) {
    return (<button
        disabled={disabled}
        onClick={onClick}
        className={"uppercase text-nowrap px-5 h-10 text-white " + (disabled ? "bg-gray " : "bg-black hover:bg-black-hover ") + className}
    >{children}</button>);
}

export function CloseButton({onClick, size = 30, className, disabled=false}: {onClick: () => void, size?: number, className?: string, disabled?: boolean}) {
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
    ><CrossIcon color={color} size={size} className={(isHover ? "stroke-2 stroke-black-hover" : "")}/></button>);
}

export function SecondaryButton({children, onClick, className} : {children: string, onClick: () => void, className?: string}) {
    return (<button
        className={"uppercase text-nowrap px-5 h-10 bg-white text-black " + className}
        onClick={onClick}
    >{children}</button>);
}

export function SmallButton({children, onClick, className} : {children: string, onClick: () => void, className?: string}) {
    return (<button
        className={"text-sm text-gray hover:underline hover:underline-gray " + className}
        onClick={onClick}
    >{children}</button>);
}