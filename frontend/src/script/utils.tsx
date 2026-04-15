import {useState} from "react";
import {tmovie} from "@/types/movie";

export function useRandomBackdrop(movie: tmovie) {
    const [randomBackdrop] = useState(() => {
        const index = Math.floor(Math.random() * movie.backdrops.length);
        return movie.backdrops[index];
    });

    return (randomBackdrop);
}
