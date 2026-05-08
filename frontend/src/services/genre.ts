"use client"

import {iGenre} from "@/types/movie";

export type tResponseGenre = {genres: iGenre[]};

export async function fetchGenres(language: string): Promise<tResponseGenre> {
    const url = `https://api.themoviedb.org/3/genre/movie/list?language=${language}`;

    const response = await fetch(
        url,
        {
            method: 'GET',
            headers: {
                accept: 'application/json',
                Authorization: `Bearer ${process.env.NEXT_PUBLIC_TMDB_API_KEY}`
            }
        }
    );

    if (!response.ok)
        throw new Error("Erreur API");
    return response.json();
}
