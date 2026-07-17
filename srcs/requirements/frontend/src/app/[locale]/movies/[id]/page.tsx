"use client";

import React, {useEffect, useState} from "react";
import {useParams} from "next/navigation";
import {startTorrentStreaming, useMovie, useTorrents} from "@/services/movies.service";
import useHandleError from "@/hooks/useHandleError";
import {ApiError} from "@/services/ApiError";
import MovieHero from "@/components/features/movie/MovieHero";
import MovieInfoSection from "@/components/features/movie/MovieInfoSection";
import CommentsSection from "@/components/features/comment/CommentsSection";
import getBestTorrent from "@/utils/getBestTorrent";
import useNotification from "@/contexts/NotificationContext";
import {useTranslations} from "next-intl";
import {useUserFilmHistory} from "@/services/users.service";
import addProgressToMovie from "@/utils/addProgressToMovie";
import useAuth from "@/contexts/AuthContext";
import {iMovieDetails} from "@/types/movie";

export default function MoviePage() {
    const params = useParams();
    const id = params.id as string;
    const {user} = useAuth();
    const {data, error} = useMovie(id);
    const {data: watchHistory} = useUserFilmHistory(user?.id);
    const movie = addProgressToMovie(watchHistory?.data, data?.data) as iMovieDetails;
    const [errorNode, setErrorNode] = useState<React.ReactNode>(null);
    const handleError = useHandleError();
    const [torrentId, setTorrentId] = useState<string | undefined>();
    const [startVideo, setStartVideo] = useState(false);
    const {data: torrents} = useTorrents(data?.data?.imdb_id)
    const {addNotification} = useNotification();
    const tError = useTranslations("notifications.error");

    useEffect(() => {
        if (error) {
            const node = handleError(error as ApiError, "Film");
            // eslint-disable-next-line react-hooks/set-state-in-effect
            setErrorNode(node);
        }
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [error]);

    useEffect(() => {
        const startDownloading = async () => {
            if (torrentId) {
                try {
                    await startTorrentStreaming(torrentId).then(() => {
                        setStartVideo(true);
                    });
                } catch (error) {
                    if (error instanceof ApiError)
                        addNotification(error.notificationMsg, "error");
                    else
                        addNotification(tError("unknown"), "error");
                    setTorrentId(undefined);
                }
            }
        }

        startDownloading().then(() => {});
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [torrentId]);

    if (errorNode)
        return (errorNode);

    const handleTorrent = async () => {
        const selectedTorrent = getBestTorrent(torrents?.data);
        if (torrents && selectedTorrent)
            setTorrentId(selectedTorrent.id);
        else
            addNotification(tError("torrentNotFound"), "error");
    }

    return (<div className="flex flex-col gap-4 sm:gap-6 xl:gap-10">
        <MovieHero movie={movie} onClick={torrents ? handleTorrent : undefined} torrentId={torrentId} startVideo={startVideo} torrents={torrents?.data} setTorrentId={setTorrentId} />
        <MovieInfoSection movie={data?.data}/>
        {data ? <CommentsSection movie={data.data}/> : <div/>}
    </div>);
}
