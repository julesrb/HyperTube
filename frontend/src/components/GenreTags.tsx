import React, {Dispatch, SetStateAction} from "react";
import {useModal} from "@/context/ModalContext";
import {useRouter} from "@/i18n/navigation";
import {useGenres} from "@/context/useGenres";
import {iGenre} from "@/types/movie";
import {useLocale} from "next-intl";
import {tLocale} from "@/i18n/request";

export default function GenreTags({genreIds, genreCount, className = "", limit, setFilterGenre}: { genreIds?: number[], genreCount?: number, className?: string, limit?: number, setFilterGenre?: Dispatch<SetStateAction<iGenre[]>>}) {
    let addLimit = false;
    const {openModal, closeModal} = useModal();

    const locale = useLocale() as tLocale;
    const {data, isLoading, error} = useGenres(locale);

    if (!data?.genres)
        return <div>Loading...</div>;
    if (isLoading)
        return <div>Loading...</div>; // todo remake

    let showGenres = data.genres;
    if (error)
        return <div>Error</div>;

    if (genreIds)
        showGenres = data.genres.filter(g => genreIds.includes(g.id));
    else if (limit !== undefined && data.genres.length > limit) {
        addLimit = true;
        showGenres = data.genres.slice(0, limit);
    }
    else if (genreCount)
        showGenres = data.genres.slice(0, genreCount); // todo select random genre

    return (<div className={"flex gap-2 sm:gap-4 flex-wrap " + className}>
        {showGenres.map((genre) => (<GenreTag key={genre.id} closeModal={closeModal} setFilterGenre={setFilterGenre}>{genre}</GenreTag>))}
        {addLimit && <button className="relative right-2 font-8xl hover:underline" onClick={() => {
            openModal({type: "genre", genres: genreIds, setFilterGenre: setFilterGenre});
        }}>...</button>}
    </div>);
}

function GenreTag({children, closeModal, setFilterGenre}: { children: iGenre, closeModal?: () => void, setFilterGenre?: Dispatch<SetStateAction<iGenre[]>> }) {
    const router = useRouter();

    const handleClick = () => {
        if (setFilterGenre)
            setFilterGenre((prev: iGenre[]) => {
                if (!prev.includes(children))
                    return [...prev, children];
                return prev;
            });
        else
            router.push(`/movies?genre=${children.id}`);
        if (closeModal)
            closeModal();
    }
    return (<button onClick={handleClick} className="text-nowrap px-3 custom-condensed border text-2xl custom-btn">
        {children.name}
    </button>);
}
