"use client";

import {iMovie} from "@/types/movie";
import React, {useEffect, useMemo, useState} from "react";
import {HypertubeLogo} from "@/components/Icons";
import {useTranslations} from "next-intl";
import Colors from "@/components/Colors";
import useAuth from "@/contexts/AuthContext";
import useResponsiveSize from "@/hooks/useResponsiveSize";
import {useMovies} from "@/services/movies.service";
import MoviesHero from "@/components/features/movie/MoviesHero";
import GenreTags from "@/components/features/genre/GenreTags";
import Section from "@/components/ui/Section";
import MoviesGrid from "@/components/features/movie/MoviesGrid";
import useModal from "@/contexts/ModalContext";
import shuffleArray from "@/utils/shuffleArray";
import {useUserFilmHistory} from "@/services/users.service";

export default function HomePage() {
    const {user} = useAuth();
    const t = useTranslations("home");
    const size = useResponsiveSize();
    const genreCount = {xs: 3, md: 4, lg: 6, xl: 8}[size];
    const heightAnimationLogo = {xs: 100, md: 200, lg: 250, xl: 300}[size];

    const {data: continueWatchingData} = useUserFilmHistory(user?.id);
    const {data: movies} = useMovies();
    const popular = filterAlreadyWatch(continueWatchingData?.data, movies?.data);
    const shuffledPopular = useMemo(() => shuffleArray(popular), [popular])

    const {data: featuredMovies} = useMovies("featured");
    const featured = filterAlreadyWatch(continueWatchingData?.data, featuredMovies?.data);
    const shuffledFeatured = useMemo(() => shuffleArray(featured), [featured])

    const mostRated = filterAlreadyWatch(continueWatchingData?.data, featuredMovies && movies ? [...featuredMovies.data, ...movies.data] : undefined).filter((m) => m.note > 7);
    const shuffledMostRated = useMemo(() => shuffleArray(mostRated), [mostRated])

    const continueWatching = continueWatchingData ? continueWatchingData.data.filter((m) => !m.complete) : [];

    const {data: dirctedWatchMovies} = useMovies("directstream", undefined, !!user);
    const filterDirectedWatchMovies = filterAlreadyWatch(continueWatchingData?.data, dirctedWatchMovies?.data);

    return (<div>
        <AnimateLogo maxHeight={heightAnimationLogo} />
        <MoviesHero movies={(shuffledFeatured).slice(0, 5)}/>
        <GenreTags genreCount={genreCount} className="justify-center w-full my-6 md:my-8"/>

        <div className="flex flex-col gap-4 px-4 sm:gap-6 sm:px-6" >
            {(continueWatching.length > 0) &&
            <Section title={t("continueWatching")} href="/users?tab=history">
                <MoviesGrid movieSets={continueWatching.slice(0, 3)} setLimit={true}/>
            </Section>}

            {(shuffledFeatured.length > 0) && <Section title={t("featured")} href="/movies?q=featured">
                <MoviesGrid movieSets={shuffledFeatured} setLimit={true}/>
            </Section>}

            {(shuffledPopular.length > 0) && <Section title={t("popular")} href="/movies?q=popular">
                <MoviesGrid movieSets={shuffledPopular} setLimit={true}/>
            </Section>}

            {(shuffledMostRated.length > 0) && <Section title={t("mostRated")} href="/movies?sort=most_rated">
                <MoviesGrid movieSets={shuffledMostRated} setLimit={true}/>
            </Section>}

            {filterDirectedWatchMovies.length > 0 && <Section title={t("directStream")} href="/movies?q=directstream">
                <MoviesGrid movieSets={filterDirectedWatchMovies} setLimit={true}/>
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
        window.addEventListener("wheel", handleWheel, {passive: false});
        return () => {window.removeEventListener("wheel", handleWheel);};
    }, [activeModal, maxHeight, minHeight]);

    useEffect(() => {
        // eslint-disable-next-line react-hooks/set-state-in-effect
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

function filterAlreadyWatch(watchHistory?: iMovie[], movies?: iMovie[]) {
    if (!movies)
        return [];
    if (!watchHistory || !watchHistory.length)
        return movies;
    return movies.filter(m => !watchHistory.find(mw => mw.imdb_id == m.imdb_id));
}
