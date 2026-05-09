import {iMovie, iMovieDetails} from "@/types/movie";
import {apiFetch, tListResponse, tResponse} from "@/services/api";

export function getMovies(locale: string, search_title?: string, page?: number, signal?: AbortSignal) {
    let endpoint = "/movies";
    if (search_title === "directstream")
        endpoint += "/directstream"
    else if (search_title)
        endpoint += `/search?title=${search_title}&page=${page}`;
    return apiFetch<tListResponse<iMovie[]>>(endpoint, locale, {signal: signal});
}

export function getWatchedMovies(locale: string) {
    return apiFetch<tListResponse<iMovie[]>>("/movies/watched", locale);
}

export function getMovie(id: string, locale: string) {
    return apiFetch<tResponse<iMovieDetails>>(`/movies/${id}`, locale);
}
