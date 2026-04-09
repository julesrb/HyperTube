import React from "react";

export default function Button({children, className}: {children: React.ReactNode, className?: string}) {
    return (<button className={"text-nowrap rounded-xs uppercase px-4 bg-purple text-white cursor-pointer border custom-btn " + className}>{children}</button>);
}
