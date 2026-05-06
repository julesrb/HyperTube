"use client";

import { useModal } from "@/context/ModalContext";
import ModalLayout from "@/components/modal/Layout";
import React, {useState} from "react";
import GenreTags from "@/components/GenreTags";
import {genres} from "@/types/genre";
import {CheckFillIcon} from "@/components/Icons";
import {Button} from "@/components/Buttons";
import {useTranslations} from "next-intl";

export function GenreModal() {
    const {activeModal, closeModal,} = useModal();
    const t = useTranslations("modal.genre");

    if (activeModal.type !== "genre" || activeModal.genres === undefined || activeModal.setFilterGenre === undefined)
        return null;

    return (<ModalLayout onClose={closeModal} title={t("title")}>
        <GenreTags genres={activeModal.genres} setFilterGenre={activeModal.setFilterGenre}/>
    </ModalLayout>);
}

export function FilterGenreModal() {
    const {activeModal, closeModal,} = useModal();
    const t = useTranslations("modal.filterGenre");
    if (activeModal.type !== "filter-genre" || activeModal.filterGenre === undefined)
        return null;
    const [filterGenre, setFilterGenre] = activeModal.filterGenre;
    // eslint-disable-next-line react-hooks/rules-of-hooks
    const [modalFilterGenre, setModalFilterGenre] = useState<string[]>(filterGenre);

    const handleSelection = (genre: string) => {
        let newGenres;
        if (modalFilterGenre.includes(genre))
            newGenres = modalFilterGenre.filter(g => g !== genre);
        else
            newGenres = [...modalFilterGenre, genre];
        setFilterGenre(newGenres);
        setModalFilterGenre(newGenres);
    }

    return (<ModalLayout onClose={closeModal} title={t("title")}>
        <div className="flex flex-col gap-2">
            {genres.map((genre) => (
                <button key={genre} className="flex gap-2" onClick={() => handleSelection(genre)}>
                    <div className={"size-5 " + (modalFilterGenre.includes(genre) ? "" : "border")}>
                        <CheckFillIcon className={modalFilterGenre.includes(genre) ? "" : "hidden"}/>
                    </div>
                    <p>{genre}</p>
                </button>))}
        </div>
        <Button className="mt-5" onClick={closeModal}>{t("apply")}</Button>
    </ModalLayout>);
}
