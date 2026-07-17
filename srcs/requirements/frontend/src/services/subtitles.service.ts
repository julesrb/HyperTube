import {ApiError} from "@/services/ApiError";
import {iDownloadSubtitle, iLoginOpenSubtitles, iSubtitle} from "@/types/subtitle";
import {tResponse} from "@/types/api";

async function fetchOpenSubtitles<T>(endpoint: string, body?: BodyInit, bearerToken?: string): Promise<T> {
    const response = await fetch(
        "https://api.opensubtitles.com/api/v1/" + endpoint,
        {
            headers: {
                "Api-Key": `${process.env.NEXT_PUBLIC_OPENSUBTITLES_API_KEY}`,
                "Accept": "application/json",
                "Content-Type": "application/json",
                "User-Agent": "<<hypertube v0.1>>",
                ...(bearerToken && {
                    Authorization: `Bearer ${bearerToken}`,
                }),
            },
            method: body ? "POST" : "GET",
            ...(body && {
                body: body
            })
        }
    );

    const data = await response.json();
    if (!response.ok)
        throw new ApiError(response.status, data);
    else
        return data;
}

export function fetchSubtitles(imdbId: string, language: string) {
    return fetchOpenSubtitles<tResponse<iSubtitle[]>>("subtitles?" + new URLSearchParams({imdb_id: imdbId, languages: language, ai_translated: "exclude"}));
}

export function loginOpenSubtitles() {
    return fetchOpenSubtitles<iLoginOpenSubtitles>("login", JSON.stringify({username: process.env.NEXT_PUBLIC_OPENSUBTITLES_USERNAME, password: process.env.NEXT_PUBLIC_OPENSUBTITLES_PASSWORD}));
}

export function downloadSubtitle(fileId?: number, bearerToken?: string) {
    return fetchOpenSubtitles<iDownloadSubtitle>("download", JSON.stringify({file_id: fileId}), bearerToken);
}
