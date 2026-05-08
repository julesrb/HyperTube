import {useEffect, useState} from "react";
type tSize = "xs" | "md" | "xl";

export function useResponsiveSize() {
    const [size, setSize] = useState<tSize>("xl");

    useEffect(() => {
        function handleResize() {
            if (window.innerWidth >= 1024)
                setSize("xl");
            else if (window.innerWidth >= 768)
                setSize("md");
            else
                setSize("xs");
        }
        handleResize();
        window.addEventListener("resize", handleResize);
        return () => window.removeEventListener("resize", handleResize);
    }, []);
    return size;
}
