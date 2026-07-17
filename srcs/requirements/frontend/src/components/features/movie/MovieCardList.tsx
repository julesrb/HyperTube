import {useRouter} from "@/i18n/navigation";
import {useTranslations} from "next-intl";
import {Dispatch, SetStateAction, useState} from "react";
import LinkLoginRequired from "@/components/ui/LinkLoginRequired";
import LoadingText from "@/components/LoadingText";
import {iMovie} from "@/types/movie";
import {iGenre} from "@/types/genre";
import {StarIcon} from "@/components/Icons";
import Button from "@/components/ui/Button/Button";
import GenreTags from "@/components/features/genre/GenreTags";
import Image from "next/image";
import {iUser} from "@/types/user";
import WatchProgress from "@/components/WatchProgress";

export default function MovieCardList({movie, user, setFilterGenre} : {movie?: iMovie, user?: iUser, setFilterGenre: Dispatch<SetStateAction<iGenre[]>>}) {
    const router = useRouter();
    const t = useTranslations("movie");
    const [isLoaded, setIsLoaded] = useState(false);

    return (<tr className="border-b">
        <td className="p-2 xl:p-4">
            <div className="border overflow-hidden aspect-3/2 relative">
                <div className="custom-noise"/>
                {!isLoaded && (<div className="custom-loading"/>)}
                {movie && <LinkLoginRequired href={"/movies/" + movie.imdb_id}>
                    <WatchProgress user={user} movie={movie} />
                    <Image
                        className={`size-full object-cover ${isLoaded ? "opacity-100" : "opacity-0"}`}
                        width={600} height={400} src={movie.backdrop_url} alt={t("posterAlt", {title: movie.title})}
                        loading="eager" onLoad={() => setIsLoaded(true)}
                    />
                    {movie.complete && <div className="custom-complete-movie"/>}
                </LinkLoginRequired>}
            </div>
        </td>
        <td className="sm:px-3">
            {movie ? <LinkLoginRequired href={"/movies/" + movie.imdb_id} className="flex gap-1 sm:gap-2 w-full">
                    <h1 className="max-w-9/10 custom-movie-title">{movie.title}</h1>
                    <span className="responsive-text-hairline">{movie.year}</span>
                </LinkLoginRequired> :
                <LoadingText/>}
        </td>
        <td/>
        <td className="hidden lg:table-cell">
            {movie && <GenreTags genreIds={movie.genres} limit={3} setFilterGenreAction={setFilterGenre}/>}
        </td>
        <td className="hidden sm:table-cell">
            <div className="flex gap-1 items-center">
                <StarIcon/>
                {movie ? <span>{movie.note.toFixed(1)}</span> :
                    <div className="h-5.5 w-5">
                        <div className="custom-loading"/>
                    </div>
                }
            </div>
        </td>
        <td className="text-right">
            <Button className="px-3" onClick={() => movie && router.push("/movies/" + movie.imdb_id)}>{t("watch")}</Button>
        </td>
    </tr>);
}
