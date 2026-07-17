import {iMovie, iMovieDetails, iProgress, iTorrent} from "@/types/movie";
import {useDebounce} from "use-debounce";
import useApiQuery, {updateTotal} from "@/hooks/useApiQuery";
import apiClient from "@/services/apiClient";
import {tListResponse, tResponse} from "@/types/api";
import {QueryClient} from "@tanstack/react-query";

function getMovie(movieId: string, locale: string) {
    return apiClient<tResponse<iMovieDetails>>(`/movies/${movieId}`, locale);
}

export function useMovie(movieId: string, enabled = true) {
    return useApiQuery(
        ["movie", movieId],
        (locale) => getMovie(movieId, locale),
        enabled,
    );
}

function getMovies(locale: string, search_title?: string, page?: number, signal?: AbortSignal) {
    let endpoint = "/movies";
    if (search_title === "directstream")
        endpoint += "/directstream"
    else if (search_title === "featured")
        endpoint += "/featured"
    else if (search_title === "watched")
        endpoint += "/watched"
    else if (search_title)
        endpoint += `/search?title=${search_title}&page=${page}`;
    return apiClient<tListResponse<iMovie[]>>(endpoint, locale, {signal});
}

export function useMovies(search_title?: string, page?: number, enabled = true) {
    const [debouncedQuery] = useDebounce(search_title ?? "", 200);

    return useApiQuery(
        ["movies", debouncedQuery, page ?? 0],
        (locale, signal) => getMovies(locale, debouncedQuery, page, signal),
        enabled
    );
}

export function updateMovieProgress(movieId: string, progress: number, pourcent: number, complete: boolean) {
    return apiClient<tListResponse<iProgress>>(`/movies/${movieId}/progress`, undefined, {method: "PATCH", body: JSON.stringify({progress, pourcent, complete})});
}

export function syncMovieProgress(queryClient: QueryClient, userId: number, movie: iMovieDetails, progress: iProgress) {
    const updatedMovie = {...movie, ...progress};
    const historyQueries = queryClient.getQueriesData<tListResponse<iMovie[]>>({queryKey: ["user-movie-history", userId]});
    historyQueries.forEach(([queryKey, current]) => {
        if (!current)
            return;
        const nextHistory = current.data.some((item) => item.imdb_id === movie.imdb_id)
            ? current.data.map((item) => item.imdb_id === movie.imdb_id ? updatedMovie : item)
            : [updatedMovie, ...current.data];
        queryClient.setQueryData(queryKey, {
            ...current,
            data: nextHistory,
            meta: updateTotal(current.meta, 1),
        });
    });
}

export function startTorrentStreaming(torrentId: string) {
    return apiClient<tListResponse<iTorrent[]>>(`/stream/${torrentId}`);
}

function getTorrents(locale: string, movieId?: string) {
    return apiClient<tListResponse<iTorrent[]>>(`/movies/${movieId}/torrents`, locale);
}

export function useTorrents(movieId?: string) {
    return useApiQuery(
        ["torrents", movieId ?? ""],
        (locale) => getTorrents(locale, movieId),
        movieId !== undefined
    );
}
