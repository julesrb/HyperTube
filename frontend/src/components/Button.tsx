import React from "react";

export default function Button({children, className}: {children: React.ReactNode, className?: string}) {
    return (<button className={"text-nowrap uppercase px-5 h-10 bg-purple text-white border border-black custom-btn " + className}>{children}</button>);
}
