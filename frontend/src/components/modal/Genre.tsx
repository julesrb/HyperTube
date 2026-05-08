"use client";

import { useModal } from "@/context/ModalContext";
import ModalLayout from "@/components/modal/Layout";
import React, {useEffect, useState} from "react";
import GenreTags from "@/components/GenreTags";
import {CheckFillIcon} from "@/components/Icons";
import {Button} from "@/components/Buttons";
import {useLocale, useTranslations} from "next-intl";
import {useGenres} from "../../context/useGenres";
import {tLocale} from "@/i18n/request";
import {iGenre} from "@/types/movie";

export function GenreModal() {
    const {activeModal, closeModal,} = useModal();
    const t = useTranslations("modal.genre");

    if (activeModal.type !== "genre" || activeModal.genres === undefined || activeModal.setFilterGenre === undefined)
        return null;

    return (<ModalLayout onClose={closeModal} title={t("title")}>
        <GenreTags genreIds ={activeModal.genres} setFilterGenre={activeModal.setFilterGenre}/>
    </ModalLayout>);
}

export function FilterGenreModal() {
    const {activeModal, closeModal,} = useModal();
    const t = useTranslations("modal.filterGenre");
    const [modalFilterGenre, setModalFilterGenre] = useState<iGenre[]>([]);
    const locale = useLocale() as tLocale;
    const {data, isLoading, error} = useGenres(locale);

    useEffect(() => {
        if (activeModal.filterGenre !== undefined) {
            // eslint-disable-next-line react-hooks/set-state-in-effect
            setModalFilterGenre(activeModal.filterGenre[0]);
        }
    }, [activeModal.filterGenre]);

    if (isLoading)
        return <div>Loading...</div>; // todo remake

    if (error)
        return <div>Error</div>;

    if (activeModal.type !== "filter-genre" || activeModal.filterGenre === undefined)
        return null;

    const handleSelection = (genre: iGenre) => {
        let newGenres;
        if (modalFilterGenre.includes(genre))
            newGenres = modalFilterGenre.filter(g => g !== genre);
        else
            newGenres = [...modalFilterGenre, genre];
        if (activeModal.filterGenre !== undefined)
            activeModal.filterGenre[1](newGenres);
        setModalFilterGenre(newGenres);
    }

    return (<ModalLayout onClose={closeModal} title={t("title")}>
        <div className="flex flex-col gap-2">
            {data?.genres.map((genre) => (
                <button key={genre.id} className="flex gap-2" onClick={() => handleSelection(genre)}>
                    <div className={"size-5 " + (modalFilterGenre.includes(genre) ? "" : "border")}>
                        <CheckFillIcon className={modalFilterGenre.includes(genre) ? "" : "hidden"}/>
                    </div>
                    <p>{genre.name}</p>
                </button>))}
        </div>
        <Button className="mt-5" onClick={closeModal}>{t("apply")}</Button>
    </ModalLayout>);
}
