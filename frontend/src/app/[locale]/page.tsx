"use client";

import {iMovie} from "@/types/movie";
import React, {useEffect, useState} from "react";
import {MoviesCard} from "@/components/MovieCard";
import MoviesHero from "@/components/MovieHero";
import {HypertubeLogo} from "@/components/Icons";
import GenreTags from "@/components/GenreTags";
import Section from "@/components/Section";
import {useAuth} from "@/context/AuthContext";
import {tUser} from "@/types/user";
import {useResponsiveSize} from "@/context/utils";
import {useTranslations} from "next-intl";
import {getMovies} from "@/services/movies";

export default function HomePage() {
    const {user} = useAuth();
    const t = useTranslations("home");
    let continueWatching;
    const size = useResponsiveSize();
    let genreCount = 3;
    if (size === "md")
        genreCount = 5;
    else if (size === "xl")
        genreCount = 7;
    let moviesCount = 3;
    if (size === "md")
        moviesCount = 4;
    else if (size === "xs")
        moviesCount = 2;
    let heightAnimationLogo = 100;
    if (size === "md")
        heightAnimationLogo = 200;
    else if (size === "xl")
        heightAnimationLogo = 300;

    const [movies, setMovies] = useState<iMovie[] | null>(null);
    const moviesSets = filterAlreadyWatch(user, movies);
    const mostRated = moviesSets ? structuredClone(moviesSets).sort((a, b) => b.rate - a.rate) : null;
    const popular = structuredClone(moviesSets);

    useEffect(() => {
        async function loadMovies() {
            try {
                const data = await getMovies();
                for (let i = 0; i < data.data.length; i++)
                    data.data[i].backdrop_url = data.data[i].backdrop_url.replace("/w500/", "/original/");
                setMovies(data.data);
            } catch (error) {
                console.error(error);
            }
        }
        loadMovies();
    }, []);

    if (user && movies) {
        continueWatching = user.watch_history
            .filter(h => h.watch_percent < 100)
            .map(m => movies.find(mSearch => mSearch.imdb_id === m.movie_id))
            .filter(m => m !== undefined);
    }

    return (<div>
        <AnimateLogo maxHeight={heightAnimationLogo} />
        {movies && <MoviesHero items={movies.slice(0, 5)} movie={movies[0]}/>}
        <GenreTags genreCount={genreCount} className="justify-center w-full my-8"/>

        {(continueWatching && continueWatching.length > 0) &&
        <Section title={t("continueWatching")} href="/users?tab=history">
            <MoviesCard movieSets={continueWatching.slice(0, moviesCount)} />
        </Section>}

        {popular && <Section title={t("popular")} href="/movies/">
            <MoviesCard movieSets={popular.slice(0, moviesCount)}/>
        </Section>}

        {mostRated && <Section title={t("mostRated")} href="/movies?sort=most_rated">
            <MoviesCard movieSets={mostRated.slice(0, moviesCount)}/>
        </Section>}

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

function AnimateLogo({maxHeight}: {maxHeight: number}) {
    const minHeight = maxHeight / 5;
    const [logoHeight, setLogoHeight] = useState(maxHeight);
    const [logoWidth, setLogoWidth] = useState(0);

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
    }, [maxHeight, minHeight]);

    useEffect(() => {
        setLogoHeight(maxHeight);
    }, [maxHeight]);

    useEffect(() => {
        function handleResize() {
            setLogoWidth(window.innerWidth);
        }
        handleResize();
        window.addEventListener("resize", handleResize);
        return () => window.removeEventListener("resize", handleResize);
    }, []);

    if (logoWidth === 0)
        return null;

    return (<div className="overflow-hidden w-full mb-4">
        <div className="flex gap-8">
            {[...Array(2)].map((_, i) => (
                <HypertubeLogo key={i} className="animate-marquee min-w-full" width={logoWidth} height={logoHeight} />
            ))}
        </div>
    </div>);
}

function filterAlreadyWatch(user: tUser | null, movies: iMovie[] | null) {
    if (!movies || !user)
        return null;
    return movies.filter(m => {
        for (let i = 0; i < user.watch_history.length; i++) {
            if (user.watch_history[i].movie_id === m.imdb_id)
                return false;
        }
        return true;
    });
}
