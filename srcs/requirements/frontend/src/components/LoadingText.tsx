import {useEffect, useState} from "react";

export default function LoadingText({center=false}) {
    const maxWidths = ["max-w-1/10", "max-w-3/10", "max-w-5/10", "max-w-7/10", "max-w-9/10"]
    const [randomWidth, setRandomWidth] = useState("max-w-1/2");

    useEffect(() => {
        // eslint-disable-next-line react-hooks/set-state-in-effect
        setRandomWidth(maxWidths[Math.floor(Math.random() * maxWidths.length)]);
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    return (<div className="h-8 sm:h-15 xl:h-20 w-full">
        <div className={"custom-loading " + randomWidth + (center ? " mx-auto" : "")} />
    </div>);
}
