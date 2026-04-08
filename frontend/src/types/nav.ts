import React from "react";

export type NavItem = {
    name: string
    icon: ({color, size}: {
        color?: string
        size?: number
    }) => React.JSX.Element
    href?: string
    action?: () => void
    hover?: (Icon: React.JSX.Element) => React.JSX.Element
};
