import React from "react";

export type NavItem =
    | {
    name: string;
    icon: string;
    href: string;
}
    | {
    name: string;
    icon: string;
    action: () => void;
}
    | {
    name: string;
    icon: string;
    hover: (Icon: React.JSX.Element) => React.JSX.Element;
};
