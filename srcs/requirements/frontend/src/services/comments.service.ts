import useApiQuery, {addQuery, removeQuery, updateQuery} from "@/hooks/useApiQuery";
import apiClient from "@/services/apiClient";
import {tListResponse, tResponse} from "@/types/api";
import {iComment, iCommentDetails} from "@/types/comment";
import {QueryClient} from "@tanstack/react-query";
import {iMovie} from "@/types/movie";

function getComments(movieId?: string, page?: number, locale?: string) {
    let endpoint = "/comments";
    if (movieId !== undefined && page !== undefined)
        endpoint = `/movies/${movieId}/comments?page=${page}`;
    return apiClient<tListResponse<iComment[]>>(endpoint, locale);
}

function getUserComments(userId: number, page: number, locale: string) {
    return apiClient<tListResponse<iCommentDetails[]>>(`/users/${userId}/comments?page=${page}`, locale);
}

export function postComment(locale: string, movieId: string, content: string) {
    return apiClient<tResponse<iComment>>(`/movies/${movieId}/comments`, locale, {method: "POST", body: JSON.stringify({content})});
}

export function patchComment(locale: string, commentId: number, content: string) {
    return apiClient<tResponse<iComment>>(`/comments/${commentId}`, locale, {method: "PATCH", body: JSON.stringify({content})});
}

export function deleteComment(locale: string, commentId: number) {
    return apiClient<tResponse<iComment>>(`/comments/${commentId}`, locale, {method: "DELETE"});
}

export function useComments(movieId?: string, page?: number) {
    return useApiQuery(
        ["comments", movieId, page],
        (locale) => getComments(movieId, page, locale),
    );
}

export function useProfileComments(userId: number, page: number) {
    return useApiQuery(
        ["user-comments", userId, page],
        (locale) => getUserComments(userId, page, locale),
    );
}

export function addCommentCache(queryClient: QueryClient, newComment: iComment, movie: iMovie, userId: number) {
    addQuery(queryClient, ["comments", movie.imdb_id, 0], newComment);
    const newDetailComment = structuredClone(newComment as iCommentDetails);
    newDetailComment.movie = movie;
    addQuery(queryClient, ["user-comments", userId, 0], newDetailComment);
}

export function updateCommentCache(queryClient: QueryClient, newComment: iCommentDetails, userId: number) {
    updateQuery(queryClient, ["user-comments", userId], newComment);
    const {movie, ...comment}: {movie: iMovie} & iComment = newComment;
    updateQuery(queryClient, ["comments", movie.imdb_id], comment);
}

export function removeCommentCache(queryClient: QueryClient, commentId: number, movieId: string, userId: number) {
    removeQuery(queryClient, ["comments", movieId], commentId);
    removeQuery(queryClient, ["user-comments", userId], commentId);
}
