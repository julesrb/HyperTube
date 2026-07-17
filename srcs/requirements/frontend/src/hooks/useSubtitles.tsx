"use client";

import {useQuery} from "@tanstack/react-query";
import {downloadSubtitle, fetchSubtitles, loginOpenSubtitles} from "@/services/subtitles.service";

export function useSubtitles(imdbId: string, language?: string) {
    return useQuery({
        queryKey: ["subtitles", imdbId, language],
        queryFn: () => fetchSubtitles(imdbId, language ?? "de,en,fr"),
        enabled: language !== undefined
    });
}

export function useLoginOpenSubtitles() {
    return useQuery({
        queryKey: ["opensubtiles-auth"],
        queryFn: () => loginOpenSubtitles(),
    });
}

export function useDownloadSubtitle(fileId?: number, token?: string) {
    return useQuery({
        queryKey: ["dl-subtitles", fileId, token],
        queryFn: () => downloadSubtitle(fileId, token),
        enabled: fileId !== undefined && token !== undefined,
    });
}
