import {iMovie, iMovieDetails} from "@/types/movie";
import {apiFetch, tListResponse, tResponse} from "@/services/api";

export function getMovies(search_title?: string) {
    let endpoint = "/movies";
    if (search_title) {
        endpoint += "/search?title=" + search_title;
    }
    return apiFetch<tListResponse<iMovie[]>>(endpoint);
}

export function getWatchedMovies() {
    return apiFetch<tListResponse<iMovie[]>>("users/watch/movies");
}

export function getMovie(id: string) {
    return apiFetch<tResponse<iMovieDetails>>(`/movies/${id}`);
}
