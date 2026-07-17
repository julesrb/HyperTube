"use client";

import useModal from "@/contexts/ModalContext";
import {useLocale} from "next-intl";
import useGenres from "@/hooks/useGenres";
import {tLocale} from "@/i18n/request";
import {iGenre} from "@/types/genre";
import {Dispatch, SetStateAction, useEffect, useState} from "react";
import GenreTag from "@/components/features/genre/GenreTag";

export default function GenreTags({genreIds, genreCount, className="", limit, setFilterGenreAction}: {genreIds?: number[], genreCount?: number, className?: string, limit?: number, setFilterGenreAction?: Dispatch<SetStateAction<iGenre[]>>}) {
    let addLimit = false;
    const {openModal, closeModal} = useModal();

    const locale = useLocale() as tLocale;
    const {data} = useGenres(locale);
    const [randomGenres, setRandomGenres] = useState<iGenre[]>([]);

    useEffect(() => {
        // eslint-disable-next-line react-hooks/set-state-in-effect
        setRandomGenres([...(data?.genres ?? [])].sort(() => Math.random() - 0.5).slice(0, genreCount));
    }, [data, genreCount]);

    if (!data?.genres)
        return;
    let showGenres: iGenre[] = data.genres;
    if (genreIds && limit !== undefined && genreIds?.length > limit) {
        addLimit = true;
        showGenres = data.genres.slice(0, limit);
    } else if (genreIds)
        showGenres = data.genres.filter(g => genreIds.includes(g.id));
    else if (genreCount)
        showGenres = randomGenres;

    return (<div className={"flex gap-2 sm:gap-4 flex-wrap " + className}>
        {showGenres.map((genre) => (<GenreTag key={genre.id} closeModal={closeModal} setFilterGenre={setFilterGenreAction}>{genre}</GenreTag>))}
        {addLimit && <button className="relative font-8xl hover:underline" onClick={() => {
            openModal({type: "genre", genres: genreIds, setFilterGenre: setFilterGenreAction});
        }}>...</button>}
    </div>);
}
