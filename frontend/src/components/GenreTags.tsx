import React from "react";
import Link from "next/link";
import {useModal} from "@/context/ModalContext";

export default function GenreTags({genres, className = "", limit}: { genres: string[], className?: string, limit?: number }) {
    let addLimit = false;
    let showGenres = genres;
    const {openModal} = useModal();

    if (limit !== undefined && genres.length > limit) {
        addLimit = true;
        showGenres = genres.slice(0, limit);
    }

    return (<div className={"flex gap-4 " + className}>
        {showGenres.map((genre) => (<GenreTag key={genre}>{genre}</GenreTag>))}
        {addLimit && <button className="relative right-2 font-8xl hover:underline" onClick={() => openModal("genre", genres)}>...</button>}
    </div>);
}

function GenreTag({children}: { children: React.ReactNode; }) {
    return (<Link href={"/movies/"} className="text-nowrap px-3 custom-condensed border text-2xl custom-btn">
        {children}
    </Link>);
}
