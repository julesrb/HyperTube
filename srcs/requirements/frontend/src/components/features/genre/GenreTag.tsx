import {useRouter} from "@/i18n/navigation";
import {iGenre} from "@/types/genre";
import {Dispatch, SetStateAction} from "react";

export default function GenreTag({children, closeModal, setFilterGenre, selected}: {children: iGenre, closeModal?: () => void, setFilterGenre?: Dispatch<SetStateAction<iGenre[]>>, selected?: boolean}) {
    const router = useRouter();

    const handleClick = () => {
        if (setFilterGenre)
            setFilterGenre((prev: iGenre[]) => {
                if (!prev.includes(children))
                        return [...prev, children];
                if (selected === undefined)
                    return prev;
                return prev.filter(g => g !== children);
            });
        else
            router.push(`/movies?genre=${children.id}`);
        if (closeModal)
            closeModal();
    }
    return (<button onClick={handleClick} className={"text-nowrap px-3 custom-condensed border text-2xl custom-shadow-animation-s" + (selected ? " text-white bg-black" : "")}>
        {children.name}
    </button>);
}
