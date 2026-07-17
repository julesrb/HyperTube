import apiClient, {API_URL} from "@/services/apiClient";
import {iApplication, tListResponse, tResponse} from "@/types/api";
import {iToken, iUserToken, tOauthService} from "@/types/user";
import useApiQuery from "@/hooks/useApiQuery";

export function postLogin(locale: string, data: string[]) {
    return apiClient<tResponse<iUserToken>>("/auth/login", locale, {method: "POST", body: JSON.stringify({login: data[0].trim(), password: data[1]})});
}

export function postRegister(locale: string, data: string[]) {
    return apiClient<tResponse<iUserToken>>("/auth/register", locale, {method: "POST", body: JSON.stringify({email: data[0].trim(), first_name: data[1].trim(), last_name: data[2].trim(), username: data[3].trim(), password: data[4]})});
}

export function postResetPassword(locale: string, data: string[]) {
    return apiClient<tResponse<iUserToken>>("/auth/password-reset/send-email", locale, {method: "POST", body: JSON.stringify({email: data[0].trim()})});
}

export function postSetNewPassword(locale: string, data: string[], token?: string | number) {
    return apiClient<tResponse<iUserToken>>("/auth/password-reset/set-new-password", locale, {method: "POST", body: JSON.stringify({token: token, password: data[0]})});
}

export function handleOauth(oatuhCompany: tOauthService, redirect?: string) {
    let endpoint = `${API_URL}/auth/${oatuhCompany}/login`;

    if (redirect !== null)
        endpoint += `?redirect=${redirect}`;
    window.location.href = endpoint;
}

let refreshPromise: Promise<void> | null = null;

export function refreshAccessToken(locale: string) {
    const refresh_token = localStorage.getItem("refresh_token");

    if (refreshPromise)
        return refreshPromise;

    if (refresh_token) {
        refreshPromise = apiClient<tResponse<iToken>>("/auth/refresh-token", locale, {method: "POST", body: JSON.stringify({refresh_token: refresh_token})})
            .then((res) => {
                if (res)
                    localStorage.setItem("token", res.data.access_token);
                // localStorage.setItem("token", res.data.access_token);
            }).finally(() => {
                refreshPromise = null;
            });
        return refreshPromise;
    }
}

function getApplications(locale: string, idx: number) {
    return apiClient<tListResponse<iApplication[]>>(`/oauth/applications?page=${idx}`, locale);
}

export function patchApp(locale: string, data: string[], appId?: string | number) {
    const updateData: Record<string, string> = {};

    ["name", "redirect_uri"].forEach((field, index) => {
        const newValue = data[index].trim();
        if (newValue)
            updateData[field] = newValue;
    })
    return apiClient<tResponse<iApplication>>(`/oauth/applications/${appId}`, locale, {method: "PATCH", body: JSON.stringify(updateData)});
}

export function postNewApp(locale: string, data: string[]) {
    return apiClient<tResponse<iApplication>>("/oauth/applications", locale, {method: "POST", body: JSON.stringify({name: data[0].trim(), redirect_uri: data[1].trim(),})});
}

export function deleteApp(locale: string, appId: number) {
    return apiClient<tResponse<iApplication>>(`/oauth/applications/${appId}`, locale, {method: "DELETE"});
}

export function useApplications(idx: number) {
    return useApiQuery(
        ["applications", idx],
        (locale: string) => getApplications(locale, idx),
    );
}
