import { apiFetch } from "./api";

export async function getMovies() {
    return apiFetch("/movies");
}

export async function getMovie(id: string) {
    return apiFetch(`/movies/${id}`);
}

export async function searchMovies(query: string) {
    return apiFetch(`/movies/search?q=${query}`);
}
