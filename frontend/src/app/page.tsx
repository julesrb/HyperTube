"use client";

import {movies} from "@/types/movie";
import React, {useEffect, useState} from "react";
import MovieCard from "@/components/MovieCard";
import MoviesHero from "@/components/MovieHero";
import {genres} from "@/types/genre";
import GenreTag from "@/components/GenreTag";
import {HypertubeLogo} from "@/components/Icon";

export default function HomePage() {
    return (<div>
        <AnimateLogo />
        <MoviesHero items={movies.slice(0, 5)} movie={movies[0]} />
        <div className="flex gap-4 items-center justify-center w-full mt-4">
            {genres.map((genre) => <GenreTag key={genre.id}>{genre.name}</GenreTag>)}
        </div>
        <div className="grid grid-cols-3 gap-4 mt-4 px-4">
            {movies.map((movie, index) => (<MovieCard key={index} movie={movie}/>))}
        </div>

        <div className="flex w-full mt-5">
            <div className="h-4 w-full bg-yellow-hover"></div>
            <div className="h-4 w-full bg-pink-hover"></div>
            <div className="h-4 w-full bg-green-hover"></div>
            <div className="h-4 w-full bg-purple-hover"></div>
            <div className="h-4 w-full bg-blue-hover"></div>
            <div className="h-4 w-full bg-red-hover"></div>
        </div>
        <div className="flex w-full">
            <div className="h-4 w-full bg-yellow"></div>
            <div className="h-4 w-full bg-pink"></div>
            <div className="h-4 w-full bg-green"></div>
            <div className="h-4 w-full bg-purple"></div>
            <div className="h-4 w-full bg-blue"></div>
            <div className="h-4 w-full bg-red"></div>
        </div>
    </div>);
}

function AnimateLogo() {
    const maxHeight = 400;
    const minHeight = 40;
    const [logoHeight, setLogoHeight] = useState(maxHeight);

    useEffect(() => {
        let virtualScroll = 0;

        const handleWheel = (e: WheelEvent) => {
            const isAtMin = virtualScroll >= (maxHeight - minHeight);
            const isAtTop = window.scrollY === 0;

            if (!isAtMin || (isAtTop && e.deltaY < 0)) {
                e.preventDefault();

                virtualScroll += e.deltaY;
                virtualScroll = Math.max(0, Math.min(virtualScroll, maxHeight));
                setLogoHeight(maxHeight - virtualScroll);
            }
        };

        window.addEventListener("wheel", handleWheel, {passive: false,});

        return () => {window.removeEventListener("wheel", handleWheel);};
    }, []);

    return (<div className="overflow-hidden w-full mb-4">
        <div className="flex">
            {[...Array(2)].map((_, i) => (
                <HypertubeLogo key={i} className="animate-marquee min-w-full" width={window.innerWidth} height={logoHeight} />
            ))}
        </div>
    </div>);
}
