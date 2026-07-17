"use client";

import useModal from "@/contexts/ModalContext";
import {useLocale, useTranslations} from "next-intl";
import {useEffect, useState} from "react";
import {iGenre} from "@/types/genre";
import useGenres from "@/hooks/useGenres";
import {tLocale} from "@/i18n/request";
import ModalLayout from "@/components/layout/ModalLayout";
import Button from "@/components/ui/Button/Button";
import SecondaryButton from "@/components/ui/Button/SecondaryButton";
import GenreTag from "@/components/features/genre/GenreTag";

export default function FilterGenreModal() {
    const {activeModal, closeModal} = useModal();
    const t = useTranslations("modal.filterGenre");
    const [modalFilterGenre, setModalFilterGenre] = useState<iGenre[]>([]);
    const locale = useLocale() as tLocale;
    const {data} = useGenres(locale);

    useEffect(() => {
        if (activeModal.filterGenre !== undefined)
            // eslint-disable-next-line react-hooks/set-state-in-effect
            setModalFilterGenre(activeModal.filterGenre[0]);
    }, [activeModal.filterGenre]);

    useEffect(() => {
        if (activeModal.filterGenre !== undefined)
            activeModal.filterGenre[1](modalFilterGenre);
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [modalFilterGenre]);

    if (activeModal.type !== "filter-genre" || activeModal.filterGenre === undefined)
        return null;

    return (<ModalLayout onCloseAction={closeModal} title={t("title")}>
        <div className="grid grid-cols-3 gap-2">
            {data?.genres.map(genre => <GenreTag key={genre.id} selected={modalFilterGenre.includes(genre)} setFilterGenre={setModalFilterGenre}>{genre}</GenreTag>)}
        </div>
        <div className="flex">
            <Button className="mt-5" onClick={closeModal}>{t("apply")}</Button>
            <SecondaryButton className="mt-5" onClick={() => {
                closeModal();
                if (activeModal.filterGenre !== undefined)
                    activeModal.filterGenre[1]([]);
            }}>{t("reset")}</SecondaryButton>
        </div>
    </ModalLayout>);
}
