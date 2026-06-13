"use client";

import {iMovie} from "@/types/movie";
import React, {useEffect, useState} from "react";
import {HypertubeLogo} from "@/components/Icons";
import {iUser} from "@/types/user";
import {useTranslations} from "next-intl";
import Colors from "@/components/Colors";
import useAuth from "@/contexts/AuthContext";
import {useResponsiveSize} from "@/hooks/useResponsiveSize";
import {useMovies} from "@/services/movies.service";
import MoviesHero from "@/components/features/movie/MoviesHero";
import GenreTags from "@/components/features/genre/GenreTags";
import Section from "@/components/ui/Section";
import MoviesCard from "@/components/features/movie/MoviesCard";
import useModal from "@/contexts/ModalContext";

export default function HomePage() {
    const {user} = useAuth();
    const t = useTranslations("home");
    let continueWatching;
    const size = useResponsiveSize();

    let genreCount = 3;
    if (size === "md")
        genreCount = 4;
    else if (size === "lg")
        genreCount = 6;
    else if (size === "xl")
        genreCount = 8;

    let heightAnimationLogo = 100;
    if (size === "md")
        heightAnimationLogo = 200;
    else if (size === "lg")
        heightAnimationLogo = 250;
    else if (size === "xl")
        heightAnimationLogo = 300;

    const {data: movies} = useMovies();
    const {data: featuredMovies} = useMovies("featured");
    const {data: allDirctedWatchMovies} = useMovies("directstream", undefined, !!user);
    const featured = filterAlreadyWatch(user, featuredMovies?.data);
    const dirctedWatchMovies = filterAlreadyWatch(user, allDirctedWatchMovies?.data);
    const moviesSets = filterAlreadyWatch(user, featuredMovies && movies ? [...featuredMovies.data, ...movies.data] : undefined);
    const mostRated = moviesSets ? moviesSets.filter((film) => film.note > 7) : null;
    const popular = filterAlreadyWatch(user, movies?.data);

    if (user && movies) {
        continueWatching = user.watch_history
            .filter(h => h.watch_percent < 100)
            .map(m => movies.data.find(mSearch => mSearch.imdb_id === m.movie_id))
            .filter(m => m !== undefined);
    }

    const shuffleArray = (array: iMovie[] | null) => {
        if (array === null)
            return [];
        const shuffled = [...array];
        for (let i = shuffled.length - 1; i > 0; i--) {
            // eslint-disable-next-line react-hooks/purity
            const j = Math.floor(Math.random() * (i + 1));
            [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
        }
        return shuffled;

    };

    return (<div>
        <AnimateLogo maxHeight={heightAnimationLogo} />
        <MoviesHero movies={shuffleArray(featured).slice(0, 5)} />
        <GenreTags genreCount={genreCount} className="justify-center w-full my-6 md:my-8"/>

        <div className="flex flex-col gap-4 px-4 sm:gap-6 sm:px-28" >
            {(continueWatching && continueWatching.length > 0) &&
            <Section title={t("continueWatching")} href="/users?tab=history">
                <MoviesCard movieSets={continueWatching} setLimit={true} />
            </Section>}

            {(featured && featured.length > 0) && <Section title={t("featured")} href="/movies?q=featured">
                <MoviesCard movieSets={shuffleArray(featured)} setLimit={true} />
            </Section>}

            {(popular && popular.length > 0) && <Section title={t("popular")} href="/movies?q=popular">
                <MoviesCard movieSets={shuffleArray(popular)} setLimit={true} />
            </Section>}

            {(mostRated && mostRated.length > 0) && <Section title={t("mostRated")} href="/movies?sort=most_rated">
                <MoviesCard movieSets={shuffleArray(mostRated)} setLimit={true} />
            </Section>}

            {dirctedWatchMovies && dirctedWatchMovies.length > 0 && <Section title={t("directStream")} href="/movies?q=directstream">
                <MoviesCard movieSets={dirctedWatchMovies} setLimit={true} />
            </Section>}
        </div>

        <Colors className="mt-4 sm:mt-6" />
    </div>);
}

function AnimateLogo({maxHeight}: {maxHeight: number}) {
    const {activeModal} = useModal();
    const minHeight = maxHeight / 5;
    const [logoHeight, setLogoHeight] = useState(maxHeight);
    const [logoWidth, setLogoWidth] = useState(0);

    useEffect(() => {
        if (activeModal.type !== null)
            return ;
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
    }, [activeModal, maxHeight, minHeight]);

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

function filterAlreadyWatch(user: iUser | null, movies?: iMovie[]) {
    if (!movies)
        return null;
    if (!user)
        return movies;
    return movies.filter(m => {
        for (let i = 0; i < user.watch_history.length; i++) {
            if (user.watch_history[i].movie_id === m.imdb_id)
                return false;
        }
        return true;
    });
}
