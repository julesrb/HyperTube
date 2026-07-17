import React from "react";

export default function Join({items}: {items: string[]}) {
    return (<p className="inline">
        {items.map((i, index) => (<span key={index}>
                <button className="custom-underline">{i}</button>
            {index < items.length - 1 && " ,   "}
            </span>))}
    </p>);
}
