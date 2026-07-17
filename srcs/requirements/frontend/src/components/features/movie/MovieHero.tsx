import { useTranslations } from "next-intl";
import React, {useState} from "react";
import {iMovie, iMovieDetails, iTorrent} from "@/types/movie";
import Image from "next/image";
import LinkLoginRequired from "@/components/ui/LinkLoginRequired";
import SecondaryButton from "@/components/ui/Button/SecondaryButton";
import useAuth from "@/contexts/AuthContext";
import VideoPlayer from "@/components/ui/VideoPlayer";
import { API_URL } from "@/services/apiClient";
import SmallText from "@/components/ui/SmallText";
import useModal from "@/contexts/ModalContext";
import Button from "@/components/ui/Button/Button";

export default function MovieHero({movie, onClick, onSlide, torrentId, startVideo, torrents, setTorrentId}: {movie?: iMovie | iMovieDetails; onClick?: () => void; onSlide?: (side: number) => void; torrentId?: string; startVideo?: boolean, torrents?: iTorrent[], setTorrentId?: (selectTorrentId: string) => void}) {
    const t = useTranslations("movie");
    const {user} = useAuth();
    const [isLoaded, setIsLoaded] = useState(false);
    const [errorStr, setError] = useState<undefined | string>();
    const {openModal} = useModal();

    const hasVideo = !!torrentId;
    const hasMovie = !!movie;
    const canSlide = !!onSlide;

    const imageUrl = movie?.backdrop_url?.replace("/w500/", "/original/");

    const renderVideo = hasVideo && startVideo;

    const renderImage = hasMovie;

    const renderLeftSlide = !hasVideo && canSlide;
    const renderRightSlide = !hasVideo && canSlide;

    const renderClickableLayer = !hasVideo && (
        onClick ? (<div className="size-full z-20 absolute custom-cursor-play" onClick={onClick} />) :
            (movie && (<LinkLoginRequired href={`/movies/${movie.imdb_id}`} className="h-full w-full z-20 absolute"/>)
        )
    );

    const handleRightClick = (e?: React.MouseEvent<HTMLDivElement>) => {
        e?.preventDefault();
        openModal({type: "select-torrent", torrents: torrents, setTorrentId: setTorrentId});
    };

    const renderTorrentError = () => {
        return (<div className="size-full absolute inset-0 bg-black/80 flex items-center justify-center overflow-hidden">
            <div className="max-w-4/5 sm:max-w-130 bg-white border p-4 sm:p-8 shadow-2xl text-center space-y-2 sm:space-y-4">
                <p className="text-sm sm:text-xl font-semibold text-red">{t("torrentError")}</p>
                <SmallText>{errorStr}</SmallText>
                <Button onClick={handleRightClick}>{t("chooseAnotherTorrent")}</Button>
            </div>
        </div>);
    };

    const renderContent = () => {
        if (hasVideo && !startVideo) {
            return (<div className="absolute mx-auto w-full text-center bottom-1/20 max-w-70">
                <SmallText className="my-2 xl:my-4 text-white">{t("movieDownloading")}</SmallText>
            </div>);
        }

        if (!movie || startVideo)
            return null;

        return (<LinkLoginRequired href={`/movies/${movie.imdb_id}`} className="absolute z-40 max-w-2/3 bottom-1/20">
            {onClick && !onSlide ?
                (<SecondaryButton className="my-2 xl:my-4 font-bold md:h-12" onClick={onClick} onContextMenu={handleRightClick} >{t("watch")}</SecondaryButton>)
                : (<h1 className="relative hover:underline decoration-3 underline-offset-3">{movie.title}
                    <span className="absolute -right-8 sm:-right-13 xl:-right-18 responsive-text-hairline">
                        {movie.year}
                    </span>
                </h1>
            )}
        </LinkLoginRequired>);
    };

    return (<div className="px-4 sm:px-6 min-w-full">
        <div className="relative flex flex-col items-center gap-4 aspect-video xl:aspect-21/9 border">
            {user && movie && renderVideo && !errorStr && (<VideoPlayer movie={movie as iMovieDetails} user={user} src={`${API_URL}/stream/${torrentId}/index`} setErrorAction={setError} tAction={t}/>)}

            {renderImage && imageUrl && (<Image className={"absolute inset-0 size-full object-cover " + (isLoaded ? "opacity-100" : "opacity-0")}
                    width={5000} height={5000} loading="eager"
                    onLoad={() => setIsLoaded(true)}
                    src={imageUrl} alt={t("posterAlt", { title: movie.title })}/>)}

            {renderLeftSlide && (<div className="h-full w-50 z-30 absolute left-0 custom-cursor-left" onClick={() => onSlide?.(-1)}/>)}
            {renderRightSlide && (<div className="h-full w-50 z-30 absolute right-0 custom-cursor-right" onClick={() => onSlide?.(1)}/>)}
            {renderClickableLayer}
            <div className="absolute inset-0 text-white flex items-end justify-center text-center mx-auto">
                <div className="custom-noise" />
                <div className={isLoaded ? "bg-gradient" : "custom-loading"} />
                {errorStr && renderTorrentError()}
                {hasVideo && !errorStr && <div className="custom-loading-dark opacity-80" />}
                {renderContent()}
            </div>
        </div>
    </div>);
}
