import {useState} from "react";
import {tmovie} from "@/types/movie";

export function useRandomBackdrop(movie: tmovie) {
    const [randomBackdrop] = useState(() => {
        if (!movie.backdrops?.length) return null;

        const index = Math.floor(Math.random() * movie.backdrops.length);
        return movie.backdrops[index];
    });

    return (randomBackdrop);
}
