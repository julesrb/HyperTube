"use client";

import { useModal } from "@/context/ModalContext";
import ModalLayout from "@/components/modal/Layout";
import React from "react";
import GenreTags from "@/components/GenreTags";

export default function GenreModal() {
    const {activeModal, closeModal,} = useModal();

    if (activeModal.type !== "genre" || activeModal.data === undefined)
        return null;

    return (
        <ModalLayout onClose={closeModal} title="film genres">
            <GenreTags genres={activeModal.data}/>
        </ModalLayout>
    );
}
