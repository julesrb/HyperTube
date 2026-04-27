"use client";

import {movies, tMovie} from "@/types/movie";
import React, {useEffect, useState} from "react";
import {MoviesCard} from "@/components/MovieCard";
import MoviesHero from "@/components/MovieHero";
import {genres} from "@/types/genre";
import {HypertubeLogo} from "@/components/Icons";
import GenreTags from "@/components/GenreTags";
import Section from "@/components/Section";
import {useAuth} from "@/context/AuthContext";
import {tUser} from "@/types/user";

export default function HomePage() {
    const {user} = useAuth();
    const moviesSets = user ? filterAlreadyWatch(user, movies) : movies;
    const popular = structuredClone(moviesSets);
    const mostRated = structuredClone(moviesSets).sort((a, b) => b.rate - a.rate);
    let continueWatching;

    if (user) {
        continueWatching = user.watch_history
            .filter(h => h.watch_percent < 100)
            .map(m => movies.find(mSearch => mSearch.id === m.movie_id))
            .filter(m => m !== undefined);
    }
    return (<div>
        <AnimateLogo />
        <MoviesHero items={movies.slice(0, 5)} movie={movies[0]} />
        <GenreTags genres={genres} className="items-center justify-center w-full mt-4"/>

        {continueWatching &&
        <Section title="Continue to watch" href="/users?tab=history">
            <MoviesCard movieSets={continueWatching.slice(0, 3)} />
        </Section>}

        <Section title="Popular" href="/movies/">
            <MoviesCard movieSets={popular.slice(0, 3)} />
        </Section>

        <Section title="Most rated" href="/movies?sort=most_rated">
            <MoviesCard movieSets={mostRated.slice(0, 3)} />
        </Section>

        <div className="flex w-full">
            <div className="h-4 w-full bg-yellow hover:bg-yellow-hover"></div>
            <div className="h-4 w-full bg-pink hover:bg-pink-hover"></div>
            <div className="h-4 w-full bg-green hover:bg-green-hover"></div>
            <div className="h-4 w-full bg-purple hover:bg-purple-hover"></div>
            <div className="h-4 w-full bg-blue hover:bg-blue-hover"></div>
            <div className="h-4 w-full bg-red hover:bg-red-hover"></div>
        </div>
    </div>);
}

function AnimateLogo() {
    const maxHeight = 300;
    const minHeight = 50;
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
        <div className="flex gap-8">
            {[...Array(2)].map((_, i) => (
                <HypertubeLogo key={i} className="animate-marquee min-w-full" width={window.innerWidth} height={logoHeight} />
            ))}
        </div>
    </div>);
}

function filterAlreadyWatch(user: tUser, movies: tMovie[]) {
    return movies.filter(m => {
        for (let i = 0; i < user.watch_history.length; i++) {
            if (user.watch_history[i].movie_id === m.id)
                return false;
        }
        return true;
    });
}
