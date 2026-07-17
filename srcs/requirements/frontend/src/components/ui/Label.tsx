import React from "react";

export default function Label({children}: {children: string}) {
    return (<span className="font-bold break-all text-sm sm:text-base capitalize">{children}</span>)
}
