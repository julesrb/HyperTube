import React, {Dispatch, SetStateAction} from "react";
import {useModal} from "@/context/ModalContext";
import {useRouter} from "next/navigation";

export default function GenreTags({genres, className = "", limit, setFilterGenre}: { genres: string[], className?: string, limit?: number, setFilterGenre?: Dispatch<SetStateAction<string[]>>}) {
    let addLimit = false;
    let showGenres = genres;
    const {openModal, closeModal} = useModal();

    if (limit !== undefined && genres.length > limit) {
        addLimit = true;
        showGenres = genres.slice(0, limit);
    }

    return (<div className={"flex gap-2 sm:gap-4 flex-wrap " + className}>
        {showGenres.map((genre) => (<GenreTag key={genre} closeModal={closeModal} setFilterGenre={setFilterGenre}>{genre}</GenreTag>))}
        {addLimit && <button className="relative right-2 font-8xl hover:underline" onClick={() => {
            openModal({type: "genre", genres: genres, setFilterGenre: setFilterGenre});
        }}>...</button>}
    </div>);
}

function GenreTag({children, closeModal, setFilterGenre}: { children: string, closeModal?: () => void, setFilterGenre?: Dispatch<SetStateAction<string[]>> }) {
    const router = useRouter();

    const handleClick = () => {
        if (setFilterGenre)
            setFilterGenre((prev: string[]) => {
                if (!prev.includes(children))
                    return [...prev, children];
                return prev;
            });
        else
            router.push("/movies?genre=" + children);
        if (closeModal)
            closeModal();
    }
    return (<button onClick={handleClick} className="text-nowrap px-3 custom-condensed border text-2xl custom-btn">
        {children}
    </button>);
}
