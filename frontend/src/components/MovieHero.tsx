import React, {useState} from "react";
import Image from "next/image";
import {tmovie} from "@/types/movie";
import {useRandomBackdrop} from "@/script/utils";
import Link from "next/link";
import Button from "@/components/Button";

export default function MoviesHero({items, movie, onClick}: { items?: tmovie[] | string[], movie: tmovie, onClick?: () => void }) {
    const [index, setIndex] = useState(0);
    if (items === undefined)
        items = movie.backdrops;

    const slideLeft = () => setIndex((prev) => (prev - 1 + items.length) % items.length);
    const slideRight = () => setIndex((prev) => (prev + 1) % items.length);

    return (<div className="overflow-hidden w-full">
        <div className="flex transition-transform duration-600 ease-out"
             style={{transform: `translateX(-${100 * index}%)`}}>
            {items.map((item, index) => (
                <MovieHero key={index} movie={typeof item === "string" ? movie : item} backdrop={typeof item === "string" ? item : undefined} onClick={onClick} onClickLeft={slideLeft} onClickRight={slideRight}/>))}
        </div>
    </div>);
}

function MovieHero({movie, onClick, onClickLeft, onClickRight, backdrop}: { movie: tmovie, onClick?: () => void, onClickLeft: () => void, onClickRight: () => void, backdrop?: string }) {
    let randomBackdrop = useRandomBackdrop(movie);

    if (backdrop)
        randomBackdrop = backdrop;

    return (<div className="px-4 min-w-full">
        <div className="relative flex flex-col items-center gap-4 aspect-21/9 border">
            <Image className="size-full object-cover" width={5000} height={5000} loading="eager"
                   src={"/images/" + randomBackdrop} alt={"poster of movie " + movie.title}/>
            <div className="h-full w-50 z-30 absolute left-0 custom-cursor-left"
                 onClick={onClickLeft}></div>
            {onClick && <div className="h-full w-full z-20 absolute custom-cursor-play" onClick={onClick}></div>}
            <div className="h-full w-50 z-30 absolute right-0 custom-cursor-right"
                 onClick={onClickRight}></div>
            <div className="absolute inset-0 text-white flex items-end justify-center text-center mx-auto">
                <div className="bg-gradient"></div>
                <Link href={"/movies/" + movie.id} className="z-10 max-w-2/3 p-6">
                    <h1 className="relative hover:underline decoration-3 underline-offset-3">{movie.title}
                        <span className="absolute -right-18 font-hairline text-3xl tracking-normal">{movie.year}</span>
                    </h1>
                    {backdrop === undefined && <Button className="my-4" onClick={() => console.log("watch movie")}>Watch</Button>}
                </Link>
            </div>
        </div>
    </div>);
}
