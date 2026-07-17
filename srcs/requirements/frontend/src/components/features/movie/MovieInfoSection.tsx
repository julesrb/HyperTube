import {iMovieDetails} from "@/types/movie";
import React from "react";
import {useTranslations} from "next-intl";
import LoadingText from "@/components/LoadingText";
import GenreTags from "@/components/features/genre/GenreTags";
import Label from "@/components/ui/Label";
import Join from "@/components/Join";

export default function MovieInfoSection({movie} : {movie?: iMovieDetails}) {
    const t = useTranslations("movie");

    const getLenght = () => {
        if (!movie)
            return "0h";

        const hours = Math.floor(movie.runtime_minutes / 60);
        const minutes = movie.runtime_minutes % 60;
        return (`${hours}h${minutes > 10 ? "" : "0"}${minutes}`);
    }

    if (!movie)
        return (<LoadingText center={true} />);

    return (<div className="flex flex-col gap-2 xl:gap-4 max-w-full md:max-w-5/6 xl:max-w-2/3 mx-3 sm:mx-auto">
        <h1 className="flex gap-1 justify-center w-full">
            <span className="max-w-8/10 custom-movie-title">{movie.title}</span>
            <span className="responsive-text-hairline">{movie.year}</span>
        </h1>
        <InfoMovie name={t("length")}>
            <p>{getLenght()}</p>
        </InfoMovie>

        <InfoMovie name={t("genre")}>
            <GenreTags genreIds={movie.genres}/>
        </InfoMovie>

        <InfoPeoplesMovie name={t("directors")} items={[movie.director]}/>

        <InfoPeoplesMovie name={t("stars")} items={movie.cast}/>

        <InfoMovie name={t("synopsis")}>
            <p>{movie.summary}</p>
        </InfoMovie>
    </div>);
}


function InfoMovie({children, name}: {children: React.ReactNode, name: string}) {
    return (<div className="flex gap-4">
        <div className="flex justify-end w-1/4 md:w-1/3 xl:w-1/2">
            <Label>{name}</Label>
        </div>
        <div className="w-3/4 md:w-2/3 xl:w-1/2">
            {children}
        </div>
    </div>);
}

function InfoPeoplesMovie({name, items}: {name: string, items: string[]}) {
    return (<InfoMovie name={name}>
        <Join items={items}/>
    </InfoMovie>);
}
