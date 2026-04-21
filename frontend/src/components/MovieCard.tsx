import Image from "next/image";
import {tMovie} from "@/types/movie";
import Link from "next/link";
import React, {Dispatch, SetStateAction} from "react";
import {Button} from "@/components/Button";
import {StarIcon} from "@/components/Icon";
import GenreTags from "@/components/GenreTags";
import {useRouter} from "next/navigation";


// todo useRandomBackdrop ?
export function MovieCard({movie} : {movie: tMovie}) {
    return (
        <Link href={"/movies/" + movie.id} className="relative aspect-824/560 overflow-hidden group border">
            <Image className="size-full object-cover transition-transform duration-200 group-hover:scale-103" width={1000} height={1000} src={"/images/" + movie.backdrops[0]} alt={"poster of movie: " + movie.title}/>
            <div className="absolute inset-0 p-4 flex items-end">
                <div className="bg-gradient"></div>
                <h3 className="relative text-white hover:underline decoration-2 underline-offset-3 z-10 mx-auto">{movie.title}
                    <span className="absolute -right-11 font-hairline text-lg tracking-normal">{movie.year}</span>
                </h3>
            </div>
        </Link>);
}

export function ListMovieCard({movie, setFilterGenre} : {movie: tMovie, setFilterGenre: Dispatch<SetStateAction<string[]>>}) {
    const router = useRouter();
    let title = movie.title;

    if (title.length > 20)
        title = title.slice(0, 18) + "...";

    return (<tr className="border-b group">
            <td className="w-1/5 p-4">
                <div className="border overflow-hidden">
                    <Link href={"/movies/" + movie.id}>
                        <Image className="object-cover size-full transition-transform duration-200 group-hover:scale-103" width={150} height={100} src={"/images/" + movie.backdrops[0]} alt={"poster of movie: " + movie.title}/>
                    </Link>
                </div>
            </td>
            <td className="w-2/5">
                <Link href={"/movies/" + movie.id} className="ml-2 flex gap-2 w-full">
                    <h1 className="hover:underline decoration-2 underline-offset-3 text-nowrap">{title}</h1>
                    <span className="font-hairline text-lg tracking-normal">{movie.year}</span>
                </Link>
            </td>
            <td className="w-0"></td>
            <td className="w-5/20">
                <GenreTags genres={movie.genres} limit={3} setFilterGenre={setFilterGenre}/>
            </td>
            <td className="w-1/20">
                <div className="flex gap-1 items-center">
                    <StarIcon />
                    <span className="font-medium">{movie.rate}</span>
                </div>
            </td>
            <td>
                <Button onClick={() => router.push("/movies/" + movie.id)}>watch</Button>
            </td>
    </tr>);
}
