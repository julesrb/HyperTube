"use client";

import {useParams} from "next/navigation";
import {movies} from "@/types/movie";
import React from "react";
import CommentSection from "@/app/movies/[id]/CommentSection";
import MovieInfoSection from "@/app/movies/[id]/MovieInfoSection";
import MoviesHero from "@/components/MovieHero";

export default function MoviePage() {
    const params = useParams();
    const movie = movies.find((movie) => movie.id === Number(params.id));

    if (!movie) return <p>Movie not found</p>;

    return (<div className="mx-4">
        <MoviesHero movie={movie} items={movie.backdrops} onClick={() => console.log('play movie')} />
        <MovieInfoSection movie={movie} />
        <CommentSection />
    </div>);
}
