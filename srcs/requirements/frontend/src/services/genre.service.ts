"use client"

import {deGenres, enGenres, frGenres, iGenre} from "@/types/genre";

export type tResponseGenre = {genres: iGenre[]};

export async function fetchGenres(language: string): Promise<tResponseGenre> {
    const url = `https://api.themoviedb.org/3/genre/movie/list?language=${language}`;

    try {
        const response = await fetch(url, {
            method: "GET",
            headers: {
                accept: "application/json",
                Authorization: `Bearer ${process.env.NEXT_PUBLIC_TMDB_API_KEY}`
            }
        });

        if (!response.ok)
            return getFallbackGenres(language);
        return await response.json();
    } catch {
        return getFallbackGenres(language);
    }
}

function getFallbackGenres(language: string) {
    switch (language) {
        case "fr":
            return { genres: frGenres };
        case "de":
            return { genres: deGenres };
        default:
            return { genres: enGenres };
    }
}
