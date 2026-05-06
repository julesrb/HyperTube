import Image from "next/image";
import {tMovie} from "@/types/movie";
import {Link} from "@/i18n/navigation";
import React, {Dispatch, SetStateAction} from "react";
import {Button} from "./Buttons";
import {StarIcon} from "@/components/Icons";
import GenreTags from "@/components/GenreTags";
import {useRouter} from "@/i18n/navigation";
import {useAuth} from "@/context/AuthContext";
import {tUser} from "@/types/user";
import {useTranslations} from "next-intl";


// todo useRandomBackdrop ?
export function MoviesCard({movieSets, className} : {movieSets: tMovie[], className?: string}) {
    const {user} = useAuth();
    return (<div className={"grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-2 sm:gap-4 " + className}>
        {movieSets.map((movie, index) => (<MovieCard key={index} movie={movie} user={user}/>))}
    </div>);
}

export function MovieCard({movie, user, className, showTitle = true} : {movie: tMovie, user: tUser | null, className?: string, showTitle?: boolean}) {
    let watchingPercent = 0;
    const t = useTranslations("movie");
    if (user) {
        const watchMovie = user.watch_history.find(h => h.movie_id === movie.id);
        if (watchMovie)
            watchingPercent = watchMovie.watch_percent;
    }
    return (<Link href={"/movies/" + movie.id} className={"relative aspect-10/7 overflow-hidden group border " + className}>
        <Image className="size-full object-cover transition-transform duration-200 group-hover:scale-103" width={1000} height={1000} src={"/images/" + movie.backdrops[0]} alt={t("posterAlt", {title: movie.title})}/>
        {watchingPercent > 0 && <div className={`absolute bottom-0 h-1 bg-${user ? user.color : "red"} z-10`} style={{width: `${watchingPercent}%`}}></div>}
        <div className="absolute inset-0 p-4 flex items-end">
            {watchingPercent === 100 ?
                <div className="size-full absolute inset-0 bg-black/60"></div> :
                <div className="bg-gradient"></div>
            }
            {
                showTitle &&
                <h3 className="relative text-white hover:underline decoration-2 underline-offset-3 z-10 mx-auto">{movie.title}
                    <span className="absolute -right-11 font-hairline text-lg tracking-normal">{movie.year}</span>
                </h3>
            }
        </div>
    </Link>);
}

export function ListMovieCard({movie, setFilterGenre} : {movie: tMovie, setFilterGenre: Dispatch<SetStateAction<string[]>>}) {
    const router = useRouter();
    const t = useTranslations("movie");
    let title = movie.title;

    if (title.length > 20)
        title = title.slice(0, 18) + "...";

    return (<tr className="border-b group">
            <td className="p-2 xl:p-4">
                <div className="border overflow-hidden">
                    <Link href={"/movies/" + movie.id}>
                        <Image className="object-cover size-full transition-transform duration-200 group-hover:scale-103" width={150} height={100} src={"/images/" + movie.backdrops[0]} alt={t("posterAlt", {title: movie.title})}/>
                    </Link>
                </div>
            </td>
            <td className="sm:pl-3">
                <Link href={"/movies/" + movie.id} className="flex gap-1 sm:gap-2">
                    <h1 className="hover:underline decoration-2 underline-offset-3 text-nowrap">{title}</h1>
                    <span className="responsive-text-hairline">{movie.year}</span>
                </Link>
            </td>
            <td></td>
            <td className="hidden lg:table-cell">
                <GenreTags genres={movie.genres} limit={3} setFilterGenre={setFilterGenre}/>
            </td>
            <td className="hidden sm:table-cell">
                <div className="flex gap-1 items-center">
                    <StarIcon />
                    <span className="font-medium">{movie.rate}</span>
                </div>
            </td>
            <td className="text-right">
                <Button className="px-3" onClick={() => router.push("/movies/" + movie.id)}>{t("watch")}</Button>
            </td>
    </tr>);
}
