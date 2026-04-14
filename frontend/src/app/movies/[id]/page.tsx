"use client";

import {useParams} from "next/navigation";
import {movies, tmovie} from "@/types/movie";
import Image from "next/image";
import React, {useState} from "react";
import CommentSection from "@/app/movies/[id]/CommentSection";
import MovieInfoSection from "@/app/movies/[id]/MovieInfoSection";

export default function MoviePage() {
    const params = useParams();
    const movie = movies.find((movie) => movie.id === Number(params.id));

    if (!movie) return <p>Movie not found</p>;

    return (<div className="mx-4">
        <HeroMovieSection movie={movie} />
        <MovieInfoSection movie={movie} />
        <CommentSection />
    </div>);
}

function HeroMovieSection({movie} : {movie: tmovie}) {
    const [index, setIndex] = useState(0);

    return (<div className="custom-cursor-play relative flex flex-col items-center gap-4 aspect-21/9 border">
        <Image className="size-full object-cover" width={5000} height={5000} loading="eager"
               src={"/images/" + movie.backdrops[index]} alt={"poster of movie " + movie.title}/>
        <div className="h-full w-50 z-30 absolute left-0 custom-cursor-left"
             onClick={() => setIndex((prev) => (prev - 1 + movie.backdrops.length) % movie.backdrops.length)}></div>
        <div className="h-full w-50 z-30 absolute right-0 custom-cursor-right"
             onClick={() => setIndex((prev) => (prev + 1) % movie.backdrops.length)}></div>
        <div className="h-full w-full z-20 absolute custom-cursor-play"
             onClick={() => console.log("play movie")}></div>
        <div className="absolute inset-0 text-white flex items-end justify-center text-center mx-auto">
            <div className="bg-gradient"></div>
            <div className="z-10 max-w-2/3 p-6">
                <h1 className="relative">{movie.title}
                    <span className="absolute -right-18 font-hairline text-3xl tracking-normal">{movie.year}</span>
                </h1>
            </div>
        </div>
    </div>);
}
