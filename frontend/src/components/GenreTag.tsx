import React from "react";

export default function GenreTag({children}: { children: React.ReactNode; }) {
    return (<button className="px-3 custom-condensed border text-2xl custom-btn">
        {children}
    </button>);
}
