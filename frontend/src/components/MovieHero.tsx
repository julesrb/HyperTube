import React, {useEffect, useState} from "react";
import Image from "next/image";
import {tMovie} from "@/types/movie";
import {useRandomBackdrop} from "@/script/utils";
import Link from "next/link";
import {SecondaryButton} from "@/components/Buttons";

export default function MoviesHero({items, movie, onClick}: { items?: tMovie[] | string[], movie: tMovie, onClick?: () => void }) {
    const [index, setIndex] = useState(0);
    if (items === undefined)
        items = movie.backdrops;

    const slideLeft = () => setIndex((prev) => (prev - 1 + items.length) % items.length);
    const slideRight = () => setIndex((prev) => (prev + 1) % items.length);
    useEffect(() => {
        const interval = setInterval(() => {
            setIndex((prev) => (prev + 1) % items.length);
        }, 4000);
        return () => clearInterval(interval);
    }, [items.length]);

    return (<div className="overflow-hidden w-full">
        <div className="flex transition-transform duration-600 ease-out"
             style={{transform: `translateX(-${100 * index}%)`}}>
            {items.map((item, index) => (
                <MovieHero key={index} movie={typeof item === "string" ? movie : item} backdrop={typeof item === "string" ? item : undefined} onClick={onClick} onClickLeft={slideLeft} onClickRight={slideRight}/>))}
        </div>
    </div>);
}

function MovieHero({movie, onClick, onClickLeft, onClickRight, backdrop}: { movie: tMovie, onClick?: () => void, onClickLeft: () => void, onClickRight: () => void, backdrop?: string }) {
    let randomBackdrop = useRandomBackdrop(movie);

    if (backdrop)
        randomBackdrop = backdrop;

    return (<div className="px-4 min-w-full">
        <div className="relative flex flex-col items-center gap-4 aspect-16/9 xl:aspect-21/9 border">
            <Image className="size-full object-cover" width={5000} height={5000} loading="eager"
                   src={"/images/" + randomBackdrop} alt={"poster of movie " + movie.title}/>
            <div className="h-full w-50 z-30 absolute left-0 custom-cursor-left"
                 onClick={onClickLeft}></div>
            {onClick ?
                <div className="h-full w-full z-20 absolute custom-cursor-play" onClick={onClick}></div>
                : <Link href={"/movies/" + movie.id} className="h-full w-full z-20 absolute"></Link>
            }
            <div className="h-full w-50 z-30 absolute right-0 custom-cursor-right"
                 onClick={onClickRight}></div>
            <div className="absolute inset-0 text-white flex items-end justify-center text-center mx-auto">
                <div className="bg-gradient"></div>
                <Link href={"/movies/" + movie.id} className="absolute z-40 max-w-2/3 bottom-1/20">
                    {
                        backdrop === undefined ?
                        <h1 className="relative hover:underline decoration-3 underline-offset-3">{movie.title}
                            <span className="absolute -right-8 sm:-right-13 xl:-right-18 responsive-text-hairline">{movie.year}</span>
                        </h1> :
                        <SecondaryButton className="my-2 xl:my-4 font-bold md:h-12" onClick={() => console.log("watch movie")}>Watch</SecondaryButton>
                    }
                </Link>
            </div>
        </div>
    </div>);
}
