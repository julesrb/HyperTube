import {useTranslations} from "next-intl";
import {iMovie} from "@/types/movie";
import {iUser} from "@/types/user";
import {useEffect, useState} from "react";
import LinkLoginRequired from "@/components/ui/LinkLoginRequired";
import Image from "next/image";

export default function MovieCard({movie, user, className, showTitle = true} : {movie: iMovie | null, user: iUser | null, className?: string, showTitle?: boolean}) {
    let watchingPercent = 0;
    const t = useTranslations("movie");
    if (user && movie) {
        const watchMovie = user.watch_history.find(h => h.movie_id === movie.imdb_id);
        if (watchMovie)
            watchingPercent = watchMovie.watch_percent;
    }
    const containerClass = "relative aspect-10/6 overflow-hidden";
    const [isLoaded, setIsLoaded] = useState(false);
    useEffect(() => {
        console.log("poster loaded", isLoaded);
    }, [isLoaded]);

    if (!movie) {
        return (<div className={containerClass}>
            <div className="custom-loading"/>
        </div>);
    }


    return (<LinkLoginRequired href={"/movies/" + movie.imdb_id} className={containerClass + " " + className}>
        <Image className={`size-full object-cover transition-transform duration-200 group-hover:scale-103 ${isLoaded ? "opacity-100" : "opacity-0"}`}
               width={1000} height={1000} src={movie.backdrop_url.replace("/w500/", "/w1280/")} alt={t("posterAlt", {title: movie.title})} loading="eager"
               onLoad={() => setIsLoaded(true)}
        />
        {watchingPercent > 0 && <div className={`absolute bottom-0 h-1 bg-${user ? user.color : "red"} z-10`} style={{width: `${watchingPercent}%`}}></div>}
        {!isLoaded && <div className="absolute inset-0 size-full"><div className="custom-loading" /></div>}
        <div className="absolute inset-0 p-4 flex items-end">
            <div className="custom-noise" />
            <div className={watchingPercent === 100 ? "size-full absolute inset-0 bg-black/60" : "bg-gradient"} />
            {showTitle &&
                <h3 className="relative text-white hover:underline decoration-2 underline-offset-3 z-10 mx-auto">{movie.title}
                    <span className="absolute -right-11 font-hairline text-lg tracking-normal">{movie.year}</span>
                </h3>}
        </div>
    </LinkLoginRequired>);
}
