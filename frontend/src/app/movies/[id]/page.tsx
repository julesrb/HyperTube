"use client";

import { useParams } from "next/navigation";
import {movies} from "@/types/movie";
import Image from "next/image";
import {useState} from "react";

export default function MoviePage() {
    const params = useParams();
    const movie = movies.find((movie) => movie.id === Number(params.id));
    const [index, setIndex] = useState(0);

    if (!movie)
        return <p>Movie not found</p>;

    return (<div>
        <div className="relative flex flex-col items-center gap-4 aspect-21/9 mx-4 border">
            <Image className="size-full object-cover transition-transform duration-200 group-hover:scale-103" width={1000} height={1000} src={"/images/" + movie.backdrops[index]} alt={"poster of movie: " + movie.title}/>
            <div className="absolute inset-0 text-white flex items-end justify-center p-6 max-w-2/3 text-center mx-auto">
                <div className="bg-gradient"></div>
                <h1 className="relative">{movie.title}
                    <span className="absolute -right-8 -top-3 font-hairline text-lg">{movie.year}</span>
                </h1>
            </div>
        </div>
            <div className="">
                <button onClick={() => setIndex((index + 1) % movie.backdrops.length)}>Next backdrop</button>
            </div>
    </div>);
}
