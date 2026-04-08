import React from "react";

export default function Button({children, className}: {children: React.ReactNode, className?: string}) {
    return (<button className={"uppercase px-2 bg-red text-white cursor-pointer custom-border custom-btn " + className}>{children}</button>);
}
