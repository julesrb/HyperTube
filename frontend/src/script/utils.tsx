import {useState} from "react";
import {tMovie} from "@/types/movie";

export function useRandomBackdrop(movie: tMovie) {
    const [randomBackdrop] = useState(() => {
        const index = Math.floor(Math.random() * movie.backdrops.length);
        return movie.backdrops[index];
    });

    return (randomBackdrop);
}
