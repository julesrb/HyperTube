import Image from "next/image";
import {tmovie} from "@/types/movie";
import {useState} from "react";
import Link from "next/link";


export default function MovieCard({movie} : {movie: tmovie}) {
    const [randomBackdrop] = useState(() => {
        if (!movie.backdrops?.length) return null;

        const index = Math.floor(Math.random() * movie.backdrops.length);
        return movie.backdrops[index];
    });

    return (
        <Link href={"/movies/" + movie.id} className="relative aspect-824/560 overflow-hidden group border">
            <Image className="size-full object-cover transition-transform duration-200 group-hover:scale-103" width={1000} height={1000} src={"/images/" + randomBackdrop} alt={"poster of movie: " + movie.title}/>
            <div className="absolute inset-0 p-4 flex items-end">
                <div className="bg-gradient"></div>
                <h3 className="text-white hover:underline decoration-2 underline-offset-3 z-10 mx-auto">{movie.title}</h3>
            </div>
        </Link>);
}
