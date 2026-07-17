"use client";

import useModal from "@/contexts/ModalContext";
import {useTranslations} from "next-intl";
import ModalLayout from "@/components/layout/ModalLayout";
import GenreTags from "@/components/features/genre/GenreTags";

export default function GenreModal() {
    const {activeModal, closeModal} = useModal();
    const t = useTranslations("modal.genre");

    if (activeModal.type !== "genre" || activeModal.genres === undefined || activeModal.setFilterGenre === undefined)
        return null;

    return (<ModalLayout onCloseAction={closeModal} title={t("title")}>
        <GenreTags genreIds ={activeModal.genres} setFilterGenreAction={activeModal.setFilterGenre}/>
    </ModalLayout>);
}
