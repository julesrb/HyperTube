import React, {useEffect, useState} from "react";

export default function IconButton({children, disabled, color="black", hoverColor, disabledColor, className, title, onClick}: {children: (color: string) => React.ReactNode, disabled?: boolean, color?: string, hoverColor?: string, disabledColor?: string, className?: string, title?: string, onClick?: () => void}) {
    const usedColor = (disabled && disabledColor) ? disabledColor : color;
    const [iconColor, setIconColor] = useState(usedColor);
    let usedHoverColor = hoverColor ?? "black-hover";
    if (!hoverColor) {
        if (color === "white")
            usedHoverColor = "white-loading";
        else if (color === "gray")
            usedHoverColor = "black";
        else if (color === "red")
            usedHoverColor = "red-hover";
    }

    useEffect(() => {
        // eslint-disable-next-line react-hooks/set-state-in-effect
        setIconColor(usedColor);
    }, [usedColor]);

    return (<button type="button" disabled={disabled} title={title} onClick={onClick} className={className} onMouseEnter={() => {
        if (!disabled)
            setIconColor(usedHoverColor);
    }} onMouseLeave={() => setIconColor(usedColor)}>
        {children(iconColor)}
    </button>)
}
