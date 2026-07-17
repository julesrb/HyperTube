import {iUser, iUserToken} from "@/types/user";
import useApiQuery from "@/hooks/useApiQuery";
import {tResponse} from "@/types/api";
import apiClient from "@/services/apiClient";
import {iMovie} from "@/types/movie";

function getUser(locale: string, userId: string) {
    return apiClient<tResponse<iUser>>(`/users/${userId}`, locale);
}

function getUserFilmHistory(locale: string, userId?: number) {
    return apiClient<tResponse<iMovie[]>>(`/users/${userId}/movie-history`, locale);
}

export function patchUser(locale: string, data: string[], userId?: number | string) {
    const updateData: Record<string, string> = {};

    if (data.length === 2) {
        updateData[data[0]] = data[1];
    } else {
        ["email", "first_name", "last_name", "username"].forEach((field, index) => {
            const newValue = data[index].trim();
            if (newValue)
                updateData[field] = newValue;
        })
    }
    return apiClient<tResponse<iUserToken>>(`/users/${userId}`, locale, {method: "PATCH", body: JSON.stringify(updateData)});
}

export function postNewPassword(locale: string, data: string[]) {
    return apiClient<tResponse<iUserToken>>(`/users/new-password`, locale, {method: "PATCH", body: JSON.stringify({current_password: data[0], new_password: data[1], new_password_confirm: data[2]})});
}

export function useUser(userId: string) {
    return useApiQuery(
        ["user", userId],
        (locale: string) => getUser(locale, userId),
    );
}

export function useUserFilmHistory(userId?: number) {
    return useApiQuery(
        ["user-movie-history", userId],
        (locale: string) => getUserFilmHistory(locale, userId),
        !!userId
    );
}
