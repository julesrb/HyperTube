import {apiFetch, tListResponse, tResponse} from "@/services/api";
import {iComment} from "@/types/comment";

export function getComments(filmId?: string) {
    let endpoint = "/comments";
    if (filmId !== undefined) {
        endpoint = `/movies/${filmId}/comments`;
    }
    return apiFetch<tListResponse<iComment[]>>(endpoint);
}

export function postComment(id: string) {
    return apiFetch<tResponse<iComment>>(`/movies/${id}`);
}

export function patchComment(id: string) {
    return apiFetch<tResponse<iComment>>(`/movies/${id}`);
}

export function deleteComment(id: string) {
    return apiFetch<tResponse<iComment>>(`/movies/${id}`);
}
