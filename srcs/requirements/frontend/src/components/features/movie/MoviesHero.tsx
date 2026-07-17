import React, {useCallback, useEffect, useRef, useState} from "react";
import {iMovie} from "@/types/movie";
import MovieHero from "@/components/features/movie/MovieHero";

export default function MoviesHero({movies}: {movies: iMovie[]}) {
    const [index, setIndex] = useState(0);
    const intervalRef = useRef<ReturnType<typeof setInterval> | undefined>(undefined);

    const startInterval = useCallback(() => {
        if (intervalRef.current)
            clearInterval(intervalRef.current);

        intervalRef.current = setInterval(() => {
            setIndex((prev) => (prev + 1) % movies.length);
        }, 6000);
    }, [movies.length]);

    const slide = (side: number) => {
        if (side > 0)
            setIndex((prev) => (prev + 1) % movies.length);
        else
            setIndex((prev) => (prev - 1 + movies.length) % movies.length);
        startInterval();
    };

    useEffect(() => {
        if (!movies.length)
            return;
        startInterval();
        return () => clearInterval(intervalRef.current);
    }, [movies.length, startInterval]);

    return (<div className="overflow-hidden w-full">
        <div className="flex transition-transform duration-600 ease-out"
             style={{transform: `translateX(-${100 * index}%)`}}>
            {movies.length > 0 ?
                movies.map((movie, index) => (<MovieHero key={index} movie={movie} onSlide={slide} />)) :
                <MovieHero movie={undefined} />
            }
        </div>
    </div>);
}
