import {useEffect, useState} from "react";
import {tMovie} from "@/types/movie";

export function useRandomBackdrop(movie: tMovie) {
    // const [randomBackdrop] = useState(() => {
    //     const index = Math.floor(Math.random() * movie.backdrops.length);
    //     return movie.backdrops[index];
    // });

    return (movie.backdrops[0]);
}


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
