import {tMovie} from "@/types/movie";
import GenreTags from "@/components/GenreTags";
import React from "react";

export default function MovieInfoSection({movie} : {movie: tMovie}) {
    return (<div className="flex flex-col gap-2 xl:gap-4 max-w-full md:max-w-5/6 xl:max-w-2/3 mx-3 sm:mx-auto">
    <h1 className="relative mx-auto mb-2">{movie.title}
        <span className="absolute -right-7 sm:-right-9 md:-right-13 xl:-right-18 responsive-text-hairline">{movie.year}</span>
    </h1>
    <InfoMovie name="Length">
        <p>{movie.length}</p>
    </InfoMovie>

    <InfoMovie name="Genre">
        <GenreTags genres={movie.genres}/>
    </InfoMovie>

    <InfoPeoplesMovie name="Directeor" items={movie.directors}/>

    <InfoPeoplesMovie name="Stars" items={movie.stars}/>

    <InfoMovie name="Synopsis">
        <p>{movie.synopsis}</p>
    </InfoMovie>
</div>);
}


function InfoMovie({children, name}: { children: React.ReactNode, name: string }) {
    return (<div className="flex gap-4">
        <div className={"flex justify-end w-1/4 md:w-1/3 xl:w-1/2"}>
            <span className="font-bold">{name}</span>
        </div>
        <div className="w-3/4 md:w-2/3 xl:w-1/2">
            {children}
        </div>
    </div>);
}

function InfoPeoplesMovie({name, items}: { name: string, items: string[] }) {
    return (<InfoMovie name={name}>
        <p>
            {items.map((i, index) => (<span key={index}>
                    <span className="custom-underline hover:cursor-pointer">{i}</span>
                {index < items.length - 1 && " ,   "}
                </span>))}
        </p>
    </InfoMovie>);
}
