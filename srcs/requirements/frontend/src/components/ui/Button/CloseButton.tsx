"use client";

import {CrossIcon} from "@/components/Icons";
import IconButton from "@/components/ui/Button/IconButton";

export default function CloseButton({onClickAction, size = 30, className, disabled=false}: {onClickAction: () => void, size?: number, className?: string, disabled?: boolean}) {
    return (<IconButton className={"text-nowrap " + className} disabled={disabled} onClick={onClickAction} disabledColor={"white"}>
        {(color: string) => <CrossIcon color={color} size={size} className={(color === "black-hover" ? "stroke-2 stroke-black-hover" : "")}/>}
    </IconButton>);
}
