"use client";

import {useParams} from "next/navigation";
import {movies} from "@/types/movie";
import React from "react";
import {CommentSection} from "@/components/Comments";
import MovieInfoSection from "@/app/[locale]/movies/[id]/MovieInfoSection";
import MoviesHero from "@/components/MovieHero";
import {useTranslations} from "next-intl";

export default function MoviePage() {
    const params = useParams();
    const movie = movies.find((movie) => movie.id === Number(params.id));
    const t = useTranslations("movie");

    if (!movie) return <p>{t("notFound")}</p>; // todo

    return (<div className="flex flex-col gap-4 sm:gap-6 xl:gap-10">
        <MoviesHero movie={movie} items={movie.backdrops} onClick={() => console.log('play movie')} />
        <MovieInfoSection movie={movie} />
        <CommentSection movie={movie} />
    </div>);
}
