import {useTranslations} from "next-intl";
import {iMovie} from "@/types/movie";
import {iUser} from "@/types/user";
import {useState} from "react";
import LinkLoginRequired from "@/components/ui/LinkLoginRequired";
import Image from "next/image";
import WatchProgress from "@/components/WatchProgress";

export default function MovieCard({movie, user, className, showTitle = true} : {movie?: iMovie, user?: iUser, className?: string, showTitle?: boolean}) {
    const t = useTranslations("movie");
    const containerClass = "relative aspect-10/7 overflow-hidden border";
    const [isLoaded, setIsLoaded] = useState(false);

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
        <WatchProgress user={user} movie={movie} />
        {!isLoaded && <div className="absolute inset-0 size-full"><div className="custom-loading"/></div>}
        <div className="absolute inset-0 p-4 flex items-end">
            <div className="custom-noise" />
            <div className={movie.complete ? "custom-complete-movie" : "bg-gradient"} />
            {showTitle &&
                <h3 className="pl-[8%] flex gap-1 justify-center w-full z-10 text-white">
                    <span className="max-w-8/10 custom-movie-title">{movie.title}</span>
                    <span className="responsive-text-hairline xl:text-xl text-xl">{movie.year}</span>
                </h3>}
        </div>
    </LinkLoginRequired>);
}
